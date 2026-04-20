package agrpc_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"testing"
	"time"

	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_TLSLoading_Success(t *testing.T) {
	// 1. Generate a self-signed CA cert/key
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Aegis AI Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caBits, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	require.NoError(t, err)

	caFile, err := os.CreateTemp("", "ca.pem")
	require.NoError(t, err)
	defer os.Remove(caFile.Name())
	err = pem.Encode(caFile, &pem.Block{Type: "CERTIFICATE", Bytes: caBits})
	require.NoError(t, err)
	err = caFile.Close()
	require.NoError(t, err)

	// 2. Generate a client cert/key signed by CA
	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"Aegis AI Test"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature,
	}
	clientBits, err := x509.CreateCertificate(rand.Reader, clientTemplate, caTemplate, &clientKey.PublicKey, caKey)
	require.NoError(t, err)

	certFile, err := os.CreateTemp("", "client.pem")
	require.NoError(t, err)
	defer os.Remove(certFile.Name())
	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: clientBits})
	require.NoError(t, err)
	err = certFile.Close()
	require.NoError(t, err)

	keyFile, err := os.CreateTemp("", "client.key")
	require.NoError(t, err)
	defer os.Remove(keyFile.Name())
	err = pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientKey)})
	require.NoError(t, err)
	err = keyFile.Close()
	require.NoError(t, err)

	// 3. Test NewClient with these files
	conf := agrpc.TLSConfig{
		Enable:     true,
		CAPath:     caFile.Name(),
		CertPath:   certFile.Name(),
		KeyPath:    keyFile.Name(),
		ServerName: "localhost",
	}

	client, err := agrpc.NewClient("localhost:1234", conf)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	_ = client.Close()
}
