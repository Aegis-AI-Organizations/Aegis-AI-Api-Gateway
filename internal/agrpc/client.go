package agrpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc/aegis/v2"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn                 *grpc.ClientConn
	PingService          v1.PingServiceClient
	ScanService          v1.ScanServiceClient
	VulnerabilityService v1.VulnerabilityServiceClient
	AuthService          v1.AuthServiceClient
	CompanyService       v1.CompanyServiceClient
}

// TLSConfig holds the paths to the certificates for mTLS.
type TLSConfig struct {
	Enable   bool
	CAPath   string
	CertPath string
	KeyPath  string
	ServerName string
}

// WithMetadata extracts identity claims from context and injects them into gRPC metadata.
func WithMetadata(ctx context.Context) context.Context {
	md := metadata.Pairs()

	// Extract from context (matching strongly-typed keys in middleware/auth.go)
	if userID, ok := ctx.Value(middleware.UserIDKey).(string); ok {
		md.Set("user-id", userID)
	}
	if companyID, ok := ctx.Value(middleware.CompanyIDKey).(string); ok {
		md.Set("company-id", companyID)
	}
	if role, ok := ctx.Value(middleware.RoleKey).(string); ok {
		md.Set("role", role)
	}
	if token, ok := ctx.Value(middleware.TokenKey).(string); ok {
		md.Set("authorization", "Bearer "+token)
	}

	if md.Len() > 0 {
		return metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}

func NewClient(addr string, conf TLSConfig) (*Client, error) {
	var opts []grpc.DialOption

	// Add keepalive parameters
	kpc := keepalive.ClientParameters{
		Time:                10 * 1000000000, // 10s
		Timeout:             20 * 1000000000, // 20s
		PermitWithoutStream: true,
	}
	opts = append(opts, grpc.WithKeepaliveParams(kpc))
	opts = append(opts, grpc.WithDefaultCallOptions(
		grpc.MaxCallSendMsgSize(50*1024*1024),
		grpc.MaxCallRecvMsgSize(50*1024*1024),
	))

	if conf.Enable {
		creds, err := loadTLSCredentials(conf)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:                 conn,
		PingService:          v1.NewPingServiceClient(conn),
		ScanService:          v1.NewScanServiceClient(conn),
		VulnerabilityService: v1.NewVulnerabilityServiceClient(conn),
		AuthService:          v1.NewAuthServiceClient(conn),
		CompanyService:       v1.NewCompanyServiceClient(conn),
	}, nil
}

func loadTLSCredentials(conf TLSConfig) (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := os.ReadFile(conf.CAPath)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Load client's certificate and private key
	clientCert, err := tls.LoadX509KeyPair(conf.CertPath, conf.KeyPath)
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	if conf.ServerName != "" {
		tlsConfig.ServerName = conf.ServerName
	}

	return credentials.NewTLS(tlsConfig), nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	if c.PingService == nil {
		return "", fmt.Errorf("ping service not initialized")
	}
	resp, err := c.PingService.Ping(WithMetadata(ctx), &v1.PingRequest{})
	if err != nil {
		return "", err
	}
	return resp.Message, nil
}

func (c *Client) StartScan(ctx context.Context, image string) (*v1.StartScanResponse, error) {
	if c.ScanService == nil {
		return nil, fmt.Errorf("scan service not initialized")
	}
	resp, err := c.ScanService.StartScan(WithMetadata(ctx), &v1.StartScanRequest{TargetImage: image})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetScanStatus(ctx context.Context, scanID string) (*v1.GetScanStatusResponse, error) {
	if c.ScanService == nil {
		return nil, fmt.Errorf("scan service not initialized")
	}
	return c.ScanService.GetScanStatus(WithMetadata(ctx), &v1.GetScanStatusRequest{ScanId: scanID})
}

func (c *Client) ListScans(ctx context.Context) ([]*v1.ScanDetails, error) {
	if c.ScanService == nil {
		return nil, fmt.Errorf("scan service not initialized")
	}
	resp, err := c.ScanService.ListScans(WithMetadata(ctx), &v1.ListScansRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Scans, nil
}

func (c *Client) GetScanReport(ctx context.Context, scanID string) ([]byte, error) {
	if c.ScanService == nil {
		return nil, fmt.Errorf("scan service not initialized")
	}
	resp, err := c.ScanService.GetScanReport(WithMetadata(ctx), &v1.GetScanReportRequest{ScanId: scanID})
	if err != nil {
		return nil, err
	}
	return resp.PdfData, nil
}

func (c *Client) GetVulnerabilities(ctx context.Context, scanID string) ([]*v1.Vulnerability, error) {
	if c.VulnerabilityService == nil {
		return nil, fmt.Errorf("vulnerability service not initialized")
	}
	resp, err := c.VulnerabilityService.GetVulnerabilities(WithMetadata(ctx), &v1.GetVulnerabilitiesRequest{ScanId: scanID})
	if err != nil {
		return nil, err
	}
	return resp.Vulnerabilities, nil
}

func (c *Client) GetEvidences(ctx context.Context, vulnID string) ([]*v1.Evidence, error) {
	if c.VulnerabilityService == nil {
		return nil, fmt.Errorf("vulnerability service not initialized")
	}
	resp, err := c.VulnerabilityService.GetEvidences(WithMetadata(ctx), &v1.GetEvidencesRequest{VulnerabilityId: vulnID})
	if err != nil {
		return nil, err
	}
	return resp.Evidences, nil
}

func (c *Client) WatchScanStatus(ctx context.Context, scanID string) (v1.ScanService_WatchScanStatusClient, error) {
	if c.ScanService == nil {
		return nil, fmt.Errorf("scan service not initialized")
	}
	return c.ScanService.WatchScanStatus(WithMetadata(ctx), &v1.WatchScanStatusRequest{ScanId: scanID})
}

func (c *Client) Login(ctx context.Context, email, password string) (*v1.LoginResponse, error) {
	if c.AuthService == nil {
		return nil, fmt.Errorf("auth service not initialized")
	}
	return c.AuthService.Login(WithMetadata(ctx), &v1.LoginRequest{Email: email, Password: password})
}

func (c *Client) Refresh(ctx context.Context, refreshToken string) (*v1.RefreshResponse, error) {
	if c.AuthService == nil {
		return nil, fmt.Errorf("auth service not initialized")
	}
	return c.AuthService.Refresh(WithMetadata(ctx), &v1.RefreshRequest{RefreshToken: refreshToken})
}

func (c *Client) Logout(ctx context.Context, refreshToken string) (*v1.LogoutResponse, error) {
	if c.AuthService == nil {
		return nil, fmt.Errorf("auth service not initialized")
	}
	return c.AuthService.Logout(WithMetadata(ctx), &v1.LogoutRequest{RefreshToken: refreshToken})
}

func (c *Client) GetMe(ctx context.Context) (*v1.GetMeResponse, error) {
	if c.AuthService == nil {
		return nil, fmt.Errorf("auth service not initialized")
	}
	return c.AuthService.GetMe(WithMetadata(ctx), &v1.GetMeRequest{})
}

func (c *Client) UpdateProfile(ctx context.Context, name, avatarURL string) (*v1.UpdateProfileResponse, error) {
	if c.AuthService == nil {
		return nil, fmt.Errorf("auth service not initialized")
	}
	return c.AuthService.UpdateProfile(WithMetadata(ctx), &v1.UpdateProfileRequest{
		Name:      name,
		AvatarUrl: avatarURL,
	})
}

func (c *Client) UpdateEmail(ctx context.Context, newEmail string) (*v1.UpdateEmailResponse, error) {
	if c.AuthService == nil {
		return nil, fmt.Errorf("auth service not initialized")
	}
	return c.AuthService.UpdateEmail(WithMetadata(ctx), &v1.UpdateEmailRequest{NewEmail: newEmail})
}

func (c *Client) UpdatePassword(ctx context.Context, oldPwd, newPwd string) (*v1.UpdatePasswordResponse, error) {
	if c.AuthService == nil {
		return nil, fmt.Errorf("auth service not initialized")
	}
	return c.AuthService.UpdatePassword(WithMetadata(ctx), &v1.UpdatePasswordRequest{OldPassword: oldPwd, NewPassword: newPwd})
}

func (c *Client) CreateCompany(ctx context.Context, name, ownerEmail string) (*v1.CreateCompanyResponse, error) {
	if c.CompanyService == nil {
		return nil, fmt.Errorf("company service not initialized")
	}
	return c.CompanyService.CreateCompany(WithMetadata(ctx), &v1.CreateCompanyRequest{Name: name, OwnerEmail: ownerEmail})
}

func (c *Client) ListCompanies(ctx context.Context) ([]*v1.CompanySummary, error) {
	if c.CompanyService == nil {
		return nil, fmt.Errorf("company service not initialized")
	}
	resp, err := c.CompanyService.ListCompanies(WithMetadata(ctx), &v1.ListCompaniesRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Companies, nil
}
