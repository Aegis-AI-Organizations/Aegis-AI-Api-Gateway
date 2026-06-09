package agrpc_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc/aegis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type mtlsPingServer struct {
	v1.UnimplementedPingServiceServer
}

func (mtlsPingServer) Ping(context.Context, *v1.PingRequest) (*v1.PingResponse, error) {
	return &v1.PingResponse{Message: "pong"}, nil
}

func TestClient_TLSRejectsUntrustedBrainCertificate(t *testing.T) {
	trustedCA, trustedCAKey, trustedCADER := newTestCA(t, "trusted-ca")
	untrustedCA, untrustedCAKey, _ := newTestCA(t, "untrusted-ca")
	trustedCAPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: trustedCADER})

	clientCert, clientKey := newTestCertificate(
		t,
		trustedCA,
		trustedCAKey,
		"gateway",
		[]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		nil,
	)
	serverCert, serverKey := newTestCertificate(
		t,
		untrustedCA,
		untrustedCAKey,
		"brain",
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		[]string{"localhost"},
	)

	serverPair, err := tls.LoadX509KeyPair(serverCert, serverKey)
	require.NoError(t, err)
	clientPool := x509.NewCertPool()
	require.True(t, clientPool.AppendCertsFromPEM(trustedCAPEM))

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	server := grpc.NewServer(grpc.Creds(credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{serverPair},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientPool,
		MinVersion:   tls.VersionTLS12,
	})))
	v1.RegisterPingServiceServer(server, mtlsPingServer{})
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Stop()

	caPath := writePEMFile(t, "ca.pem", "CERTIFICATE", trustedCADER)
	client, err := agrpc.NewClient(listener.Addr().String(), agrpc.TLSConfig{
		Enable:     true,
		CAPath:     caPath,
		CertPath:   clientCert,
		KeyPath:    clientKey,
		ServerName: "localhost",
	})
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = client.Ping(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "certificate")
}

func newTestCA(t *testing.T, commonName string) (*x509.Certificate, *rsa.PrivateKey, []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	return cert, key, der
}

func newTestCertificate(
	t *testing.T,
	ca *x509.Certificate,
	caKey *rsa.PrivateKey,
	commonName string,
	usages []x509.ExtKeyUsage,
	dnsNames []string,
) (string, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: commonName},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  usages,
		DNSNames:     dnsNames,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, ca, &key.PublicKey, caKey)
	require.NoError(t, err)
	return writePEMFile(t, commonName+".pem", "CERTIFICATE", der),
		writePEMFile(t, commonName+".key", "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key))
}

func writePEMFile(t *testing.T, name string, blockType string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: data}), 0o600)
	require.NoError(t, err)
	return path
}
