package agrpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc/aegis/v2"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/types"
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
	BillingService       v1.BillingServiceClient
	InternalAuthService  v1.InternalAuthServiceClient
	AgentService         v1.AgentServiceClient
}

// TLSConfig holds the paths to the certificates for mTLS.
type TLSConfig struct {
	Enable     bool
	CAPath     string
	CertPath   string
	KeyPath    string
	ServerName string
}

// WithMetadata extracts identity claims from context and injects them into gRPC metadata.
func WithMetadata(ctx context.Context) context.Context {
	md := metadata.Pairs()

	// Extract from context (matching strongly-typed keys in types package)
	if userID, ok := ctx.Value(types.UserIDKey).(string); ok {
		md.Set("user-id", userID)
	}
	if companyID, ok := ctx.Value(types.CompanyIDKey).(string); ok {
		md.Set("company-id", companyID)
	}
	if role, ok := ctx.Value(types.RoleKey).(string); ok {
		md.Set("role", role)
	}
	if token, ok := ctx.Value(types.TokenKey).(string); ok {
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
		Time:                60 * 1000000000, // 60s
		Timeout:             20 * 1000000000, // 20s
		PermitWithoutStream: false,
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
		BillingService:       v1.NewBillingServiceClient(conn),
		InternalAuthService:  v1.NewInternalAuthServiceClient(conn),
		AgentService:         v1.NewAgentServiceClient(conn),
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

func (c *Client) SearchCompanies(ctx context.Context, query string) ([]*v1.CompanySummary, error) {
	if c.CompanyService == nil {
		return nil, fmt.Errorf("company service not initialized")
	}
	authCtx := WithMetadata(ctx)
	newCtx := metadata.AppendToOutgoingContext(authCtx,
		"x-action", "list-companies",
		"x-query", query,
	)
	resp, err := c.CompanyService.ListCompanies(newCtx, &v1.ListCompaniesRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Companies, nil
}

func (c *Client) SearchUsers(ctx context.Context, query, companyID string) ([]*v1.CompanySummary, error) {
	if c.CompanyService == nil {
		return nil, fmt.Errorf("company service not initialized")
	}
	authCtx := WithMetadata(ctx)
	newCtx := metadata.AppendToOutgoingContext(authCtx,
		"x-action", "list-users",
		"x-query", query,
		"x-company-id", companyID,
	)
	resp, err := c.CompanyService.ListCompanies(newCtx, &v1.ListCompaniesRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Companies, nil
}

func (c *Client) AdminCreateUser(ctx context.Context, name, email, password, role, companyID string) (*v1.CreateCompanyResponse, error) {
	if c.CompanyService == nil {
		return nil, fmt.Errorf("company service not initialized")
	}
	authCtx := WithMetadata(ctx)
	newCtx := metadata.AppendToOutgoingContext(authCtx,
		"x-action", "create-user",
		"x-user-password", password,
		"x-user-role", role,
		"x-company-id", companyID,
	)
	return c.CompanyService.CreateCompany(newCtx, &v1.CreateCompanyRequest{
		Name:       name,
		OwnerEmail: email,
	})
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

func (c *Client) SetupPassword(ctx context.Context, invitationToken, newPassword string) (*v1.SetupPasswordResponse, error) {
	if c.AuthService == nil {
		return nil, fmt.Errorf("auth service not initialized")
	}
	return c.AuthService.SetupPassword(WithMetadata(ctx), &v1.SetupPasswordRequest{
		InvitationToken: invitationToken,
		NewPassword:     newPassword,
	})
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

func (c *Client) RemoveAvatar(ctx context.Context) (*v1.RemoveAvatarResponse, error) {
	if c.AuthService == nil {
		return nil, fmt.Errorf("auth service not initialized")
	}
	return c.AuthService.RemoveAvatar(WithMetadata(ctx), &v1.RemoveAvatarRequest{})
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

func (c *Client) OnboardCompany(ctx context.Context, companyName, ownerName, ownerEmail string) (*v1.OnboardCompanyResponse, error) {
	if c.CompanyService == nil {
		return nil, fmt.Errorf("company service not initialized")
	}
	return c.CompanyService.OnboardCompany(WithMetadata(ctx), &v1.OnboardCompanyRequest{
		CompanyName: companyName,
		OwnerName:   ownerName,
		OwnerEmail:  ownerEmail,
	})
}

func (c *Client) RotateAgentToken(ctx context.Context, companyID string) (*v1.RotateAgentTokenResponse, error) {
	if c.CompanyService == nil {
		return nil, fmt.Errorf("company service not initialized")
	}
	return c.CompanyService.RotateAgentToken(WithMetadata(ctx), &v1.RotateAgentTokenRequest{CompanyId: companyID})
}

func (c *Client) RevokeAgentToken(ctx context.Context, companyID string) (*v1.RevokeAgentTokenResponse, error) {
	if c.CompanyService == nil {
		return nil, fmt.Errorf("company service not initialized")
	}
	return c.CompanyService.RevokeAgentToken(WithMetadata(ctx), &v1.RevokeAgentTokenRequest{CompanyId: companyID})
}

func (c *Client) WatchCompanyUpdates(ctx context.Context) (v1.CompanyService_WatchCompanyUpdatesClient, error) {
	if c.CompanyService == nil {
		return nil, fmt.Errorf("company service not initialized")
	}
	return c.CompanyService.WatchCompanyUpdates(WithMetadata(ctx), &v1.WatchCompanyUpdatesRequest{})
}

func (c *Client) ListAuditLogs(ctx context.Context, limit, offset int32, companyID string) (*v1.ListAuditLogsResponse, error) {
	if c.CompanyService == nil {
		return nil, fmt.Errorf("company service not initialized")
	}
	return c.CompanyService.ListAuditLogs(WithMetadata(ctx), &v1.ListAuditLogsRequest{
		Limit:     limit,
		Offset:    offset,
		CompanyId: companyID,
	})
}

func (c *Client) GetBalance(ctx context.Context, companyID string) (*v1.GetBalanceResponse, error) {
	if c.BillingService == nil {
		return nil, fmt.Errorf("billing service not initialized")
	}
	return c.BillingService.GetBalance(WithMetadata(ctx), &v1.GetBalanceRequest{CompanyId: companyID})
}

func (c *Client) GetLedger(ctx context.Context, companyID string, limit, offset int32) (*v1.GetLedgerResponse, error) {
	if c.BillingService == nil {
		return nil, fmt.Errorf("billing service not initialized")
	}
	return c.BillingService.GetLedger(WithMetadata(ctx), &v1.GetLedgerRequest{
		CompanyId: companyID,
		Limit:     limit,
		Offset:    offset,
	})
}

func (c *Client) AdjustTokens(ctx context.Context, companyID string, amount int64, reason string) (*v1.AdjustTokensResponse, error) {
	if c.BillingService == nil {
		return nil, fmt.Errorf("billing service not initialized")
	}
	return c.BillingService.AdjustTokens(WithMetadata(ctx), &v1.AdjustTokensRequest{
		CompanyId: companyID,
		Amount:    amount,
		Reason:    reason,
	})
}

func (c *Client) PreFlightCheck(ctx context.Context, companyID string, ipCount, apiCount, webappCount int32) (*v1.PreFlightCheckResponse, error) {
	if c.BillingService == nil {
		return nil, fmt.Errorf("billing service not initialized")
	}
	return c.BillingService.PreFlightCheck(WithMetadata(ctx), &v1.PreFlightCheckRequest{
		CompanyId: companyID,
		TargetConfig: &v1.TargetConfig{
			IpCount:     ipCount,
			ApiCount:    apiCount,
			WebappCount: webappCount,
		},
	})
}

func (c *Client) GetUsageStats(ctx context.Context, companyID string, days int32) (*v1.GetUsageStatsResponse, error) {
	if c.BillingService == nil {
		return nil, fmt.Errorf("billing service not initialized")
	}
	return c.BillingService.GetUsageStats(WithMetadata(ctx), &v1.GetUsageStatsRequest{
		CompanyId: companyID,
		Days:      days,
	})
}
func (c *Client) VerifyToken(ctx context.Context, token string) (*v1.VerifyTokenResponse, error) {
	if c.InternalAuthService == nil {
		return nil, fmt.Errorf("internal auth service not initialized")
	}
	return c.InternalAuthService.VerifyToken(ctx, &v1.VerifyTokenRequest{Token: token})
}

func (c *Client) RegisterAgent(ctx context.Context, token, name string) (*v1.RegisterAgentResponse, error) {
	if c.AgentService == nil {
		return nil, fmt.Errorf("agent service not initialized")
	}
	return c.AgentService.RegisterAgent(ctx, &v1.RegisterAgentRequest{
		Token: token,
		Name:  name,
	})
}

func (c *Client) UpdateAgentStatus(ctx context.Context, agentID, status string) (*v1.UpdateAgentStatusResponse, error) {
	if c.AgentService == nil {
		return nil, fmt.Errorf("agent service not initialized")
	}
	return c.AgentService.UpdateAgentStatus(ctx, &v1.UpdateAgentStatusRequest{
		AgentId: agentID,
		Status:  status,
	})
}

func (c *Client) GetUploadLink(ctx context.Context, agentID, filename string) (*v1.GetUploadLinkResponse, error) {
	if c.AgentService == nil {
		return nil, fmt.Errorf("agent service not initialized")
	}
	return c.AgentService.GetUploadLink(ctx, &v1.GetUploadLinkRequest{
		AgentId:  agentID,
		Filename: filename,
	})
}

func (c *Client) ListAgents(ctx context.Context, companyID string) (*v1.ListAgentsResponse, error) {
	if c.AgentService == nil {
		return nil, fmt.Errorf("agent service not initialized")
	}
	return c.AgentService.ListAgents(WithMetadata(ctx), &v1.ListAgentsRequest{
		CompanyId: companyID,
	})
}

func (c *Client) GetAgentStatusSummary(ctx context.Context, companyID string) (*v1.GetAgentStatusSummaryResponse, error) {
	if c.AgentService == nil {
		return nil, fmt.Errorf("agent service not initialized")
	}
	return c.AgentService.GetAgentStatusSummary(WithMetadata(ctx), &v1.GetAgentStatusSummaryRequest{
		CompanyId: companyID,
	})
}

func (c *Client) VerifyAgentSecret(ctx context.Context, agentID, secret string) (*v1.VerifyAgentSecretResponse, error) {
	if c.AgentService == nil {
		return nil, fmt.Errorf("agent service not initialized")
	}
	return c.AgentService.VerifyAgentSecret(ctx, &v1.VerifyAgentSecretRequest{
		AgentId: agentID,
		Secret:  secret,
	})
}
