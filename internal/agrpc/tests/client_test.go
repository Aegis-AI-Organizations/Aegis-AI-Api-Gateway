package agrpc_test

import (
	"context"
	"fmt"
	"testing"

	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc/aegis/v2"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type MockPingServiceClient struct {
	mock.Mock
}

func (m *MockPingServiceClient) Ping(ctx context.Context, in *v1.PingRequest, opts ...grpc.CallOption) (*v1.PingResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*v1.PingResponse), args.Error(1)
}

type MockScanServiceClient struct {
	mock.Mock
}

func (m *MockScanServiceClient) StartScan(ctx context.Context, in *v1.StartScanRequest, opts ...grpc.CallOption) (*v1.StartScanResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*v1.StartScanResponse), args.Error(1)
}

func (m *MockScanServiceClient) GetScanStatus(ctx context.Context, in *v1.GetScanStatusRequest, opts ...grpc.CallOption) (*v1.GetScanStatusResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*v1.GetScanStatusResponse), args.Error(1)
}

func (m *MockScanServiceClient) GetScanReport(ctx context.Context, in *v1.GetScanReportRequest, opts ...grpc.CallOption) (*v1.GetScanReportResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*v1.GetScanReportResponse), args.Error(1)
}

func (m *MockScanServiceClient) ListScans(ctx context.Context, in *v1.ListScansRequest, opts ...grpc.CallOption) (*v1.ListScansResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*v1.ListScansResponse), args.Error(1)
}

func (m *MockScanServiceClient) WatchScanStatus(ctx context.Context, in *v1.WatchScanStatusRequest, opts ...grpc.CallOption) (v1.ScanService_WatchScanStatusClient, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(v1.ScanService_WatchScanStatusClient), args.Error(1)
}

type MockAgentServiceClient struct {
	mock.Mock
}

func (m *MockAgentServiceClient) RegisterAgent(ctx context.Context, in *v1.RegisterAgentRequest, opts ...grpc.CallOption) (*v1.RegisterAgentResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.RegisterAgentResponse), args.Error(1)
}

func (m *MockAgentServiceClient) UpdateAgentStatus(ctx context.Context, in *v1.UpdateAgentStatusRequest, opts ...grpc.CallOption) (*v1.UpdateAgentStatusResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.UpdateAgentStatusResponse), args.Error(1)
}

func (m *MockAgentServiceClient) GetUploadLink(ctx context.Context, in *v1.GetUploadLinkRequest, opts ...grpc.CallOption) (*v1.GetUploadLinkResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.GetUploadLinkResponse), args.Error(1)
}

func (m *MockAgentServiceClient) VerifyAgentSecret(ctx context.Context, in *v1.VerifyAgentSecretRequest, opts ...grpc.CallOption) (*v1.VerifyAgentSecretResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.VerifyAgentSecretResponse), args.Error(1)
}

type MockVulnerabilityServiceClient struct {
	mock.Mock
}

func (m *MockVulnerabilityServiceClient) GetVulnerabilities(ctx context.Context, in *v1.GetVulnerabilitiesRequest, opts ...grpc.CallOption) (*v1.GetVulnerabilitiesResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*v1.GetVulnerabilitiesResponse), args.Error(1)
}

func (m *MockVulnerabilityServiceClient) GetEvidences(ctx context.Context, in *v1.GetEvidencesRequest, opts ...grpc.CallOption) (*v1.GetEvidencesResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*v1.GetEvidencesResponse), args.Error(1)
}

type MockAuthServiceClient struct {
	mock.Mock
}

func (m *MockAuthServiceClient) Login(ctx context.Context, in *v1.LoginRequest, opts ...grpc.CallOption) (*v1.LoginResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.LoginResponse), args.Error(1)
}

func (m *MockAuthServiceClient) Refresh(ctx context.Context, in *v1.RefreshRequest, opts ...grpc.CallOption) (*v1.RefreshResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.RefreshResponse), args.Error(1)
}

func (m *MockAuthServiceClient) Logout(ctx context.Context, in *v1.LogoutRequest, opts ...grpc.CallOption) (*v1.LogoutResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.LogoutResponse), args.Error(1)
}

func (m *MockAuthServiceClient) GetMe(ctx context.Context, in *v1.GetMeRequest, opts ...grpc.CallOption) (*v1.GetMeResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.GetMeResponse), args.Error(1)
}

func (m *MockAuthServiceClient) UpdateProfile(ctx context.Context, in *v1.UpdateProfileRequest, opts ...grpc.CallOption) (*v1.UpdateProfileResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.UpdateProfileResponse), args.Error(1)
}

func (m *MockAuthServiceClient) UpdateEmail(ctx context.Context, in *v1.UpdateEmailRequest, opts ...grpc.CallOption) (*v1.UpdateEmailResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.UpdateEmailResponse), args.Error(1)
}

func (m *MockAuthServiceClient) UpdatePassword(ctx context.Context, in *v1.UpdatePasswordRequest, opts ...grpc.CallOption) (*v1.UpdatePasswordResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.UpdatePasswordResponse), args.Error(1)
}

func (m *MockAuthServiceClient) RemoveAvatar(ctx context.Context, in *v1.RemoveAvatarRequest, opts ...grpc.CallOption) (*v1.RemoveAvatarResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.RemoveAvatarResponse), args.Error(1)
}

type MockBillingServiceClient struct {
	mock.Mock
}

func (m *MockBillingServiceClient) GetBalance(ctx context.Context, in *v1.GetBalanceRequest, opts ...grpc.CallOption) (*v1.GetBalanceResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.GetBalanceResponse), args.Error(1)
}

func (m *MockBillingServiceClient) AdjustTokens(ctx context.Context, in *v1.AdjustTokensRequest, opts ...grpc.CallOption) (*v1.AdjustTokensResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.AdjustTokensResponse), args.Error(1)
}

func (m *MockBillingServiceClient) PreFlightCheck(ctx context.Context, in *v1.PreFlightCheckRequest, opts ...grpc.CallOption) (*v1.PreFlightCheckResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.PreFlightCheckResponse), args.Error(1)
}

func (m *MockBillingServiceClient) GetLedger(ctx context.Context, in *v1.GetLedgerRequest, opts ...grpc.CallOption) (*v1.GetLedgerResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.GetLedgerResponse), args.Error(1)
}

func (m *MockBillingServiceClient) GetUsageStats(ctx context.Context, in *v1.GetUsageStatsRequest, opts ...grpc.CallOption) (*v1.GetUsageStatsResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.GetUsageStatsResponse), args.Error(1)
}

type MockCompanyServiceClient struct {
	mock.Mock
}

func (m *MockCompanyServiceClient) CreateCompany(ctx context.Context, in *v1.CreateCompanyRequest, opts ...grpc.CallOption) (*v1.CreateCompanyResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.CreateCompanyResponse), args.Error(1)
}

func (m *MockCompanyServiceClient) ListCompanies(ctx context.Context, in *v1.ListCompaniesRequest, opts ...grpc.CallOption) (*v1.ListCompaniesResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.ListCompaniesResponse), args.Error(1)
}

func (m *MockCompanyServiceClient) OnboardCompany(ctx context.Context, in *v1.OnboardCompanyRequest, opts ...grpc.CallOption) (*v1.OnboardCompanyResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.OnboardCompanyResponse), args.Error(1)
}

func (m *MockCompanyServiceClient) WatchCompanyUpdates(ctx context.Context, in *v1.WatchCompanyUpdatesRequest, opts ...grpc.CallOption) (v1.CompanyService_WatchCompanyUpdatesClient, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(v1.CompanyService_WatchCompanyUpdatesClient), args.Error(1)
}

func (m *MockCompanyServiceClient) ListAuditLogs(ctx context.Context, in *v1.ListAuditLogsRequest, opts ...grpc.CallOption) (*v1.ListAuditLogsResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.ListAuditLogsResponse), args.Error(1)
}

func TestClient_Ping(t *testing.T) {
	mockPing := new(MockPingServiceClient)
	client := &agrpc.Client{
		PingService: mockPing,
	}

	mockPing.On("Ping", mock.Anything, mock.Anything).Return(&v1.PingResponse{Message: "pong"}, nil)

	msg, err := client.Ping(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "pong", msg)
}

func TestClient_ScanServices(t *testing.T) {
	mockScan := new(MockScanServiceClient)
	mockVuln := new(MockVulnerabilityServiceClient)
	client := &agrpc.Client{
		ScanService:          mockScan,
		VulnerabilityService: mockVuln,
	}

	mockScan.On("StartScan", mock.Anything, mock.Anything).Return(&v1.StartScanResponse{ScanId: "s1"}, nil)
	mockScan.On("GetScanStatus", mock.Anything, mock.Anything).Return(&v1.GetScanStatusResponse{Status: "RUNNING"}, nil)
	mockVuln.On("GetVulnerabilities", mock.Anything, mock.Anything).Return(&v1.GetVulnerabilitiesResponse{Vulnerabilities: []*v1.Vulnerability{}}, nil)

	resp, err := client.StartScan(context.Background(), "img")
	assert.NoError(t, err)
	assert.Equal(t, "s1", resp.ScanId)

	statusResp, err := client.GetScanStatus(context.Background(), "s1")
	assert.NoError(t, err)
	assert.Equal(t, "RUNNING", statusResp.Status)

	vulns, err := client.GetVulnerabilities(context.Background(), "s1")
	assert.NoError(t, err)
	assert.Len(t, vulns, 0)

	mockScan.On("GetScanReport", mock.Anything, mock.Anything).Return(&v1.GetScanReportResponse{PdfData: []byte("pdf")}, nil)
	report, err := client.GetScanReport(context.Background(), "s1")
	assert.NoError(t, err)
	assert.Equal(t, []byte("pdf"), report)

	mockScan.On("ListScans", mock.Anything, mock.Anything).Return(&v1.ListScansResponse{Scans: []*v1.ScanDetails{}}, nil)
	scans, err := client.ListScans(context.Background())
	assert.NoError(t, err)
	assert.Len(t, scans, 0)

	mockVuln.On("GetEvidences", mock.Anything, mock.Anything).Return(&v1.GetEvidencesResponse{Evidences: []*v1.Evidence{}}, nil)
	evidences, err := client.GetEvidences(context.Background(), "v1")
	assert.NoError(t, err)
	assert.Len(t, evidences, 0)
}

func TestClient_Failures(t *testing.T) {
	mockScan := new(MockScanServiceClient)
	mockVuln := new(MockVulnerabilityServiceClient)
	client := &agrpc.Client{
		ScanService:          mockScan,
		VulnerabilityService: mockVuln,
	}

	mockScan.On("StartScan", mock.Anything, mock.Anything).Return((*v1.StartScanResponse)(nil), fmt.Errorf("rpc error"))
	mockScan.On("GetScanStatus", mock.Anything, mock.Anything).Return((*v1.GetScanStatusResponse)(nil), fmt.Errorf("rpc error"))
	mockVuln.On("GetVulnerabilities", mock.Anything, mock.Anything).Return((*v1.GetVulnerabilitiesResponse)(nil), fmt.Errorf("rpc error"))

	mockScan.On("GetVulnerabilities", mock.Anything, mock.Anything).Return((*v1.GetVulnerabilitiesResponse)(nil), fmt.Errorf("rpc error")) // Unused maybe
	mockScan.On("GetScanReport", mock.Anything, mock.Anything).Return((*v1.GetScanReportResponse)(nil), fmt.Errorf("rpc error"))
	mockScan.On("ListScans", mock.Anything, mock.Anything).Return((*v1.ListScansResponse)(nil), fmt.Errorf("rpc error"))
	mockVuln.On("GetEvidences", mock.Anything, mock.Anything).Return((*v1.GetEvidencesResponse)(nil), fmt.Errorf("rpc error"))

	mockAuth := new(MockAuthServiceClient)
	client.AuthService = mockAuth
	mockAuth.On("Login", mock.Anything, mock.Anything).Return((*v1.LoginResponse)(nil), fmt.Errorf("rpc error"))
	mockAuth.On("Refresh", mock.Anything, mock.Anything).Return((*v1.RefreshResponse)(nil), fmt.Errorf("rpc error"))
	mockAuth.On("Logout", mock.Anything, mock.Anything).Return((*v1.LogoutResponse)(nil), fmt.Errorf("rpc error"))

	_, err := client.StartScan(context.Background(), "img")
	assert.Error(t, err)

	_, err = client.GetScanStatus(context.Background(), "s1")
	assert.Error(t, err)

	_, err = client.GetVulnerabilities(context.Background(), "s1")
	assert.Error(t, err)

	_, err = client.GetScanReport(context.Background(), "s1")
	assert.Error(t, err)

	_, err = client.ListScans(context.Background())
	assert.Error(t, err)

	_, err = client.GetEvidences(context.Background(), "v1")
	assert.Error(t, err)

	_, err = client.Login(context.Background(), "e", "p")
	assert.Error(t, err)

	_, err = client.Refresh(context.Background(), "r")
	assert.Error(t, err)

	_, err = client.Logout(context.Background(), "r")
	assert.Error(t, err)
}

func TestClient_AuthMethods(t *testing.T) {
	mockAuth := new(MockAuthServiceClient)
	client := &agrpc.Client{
		AuthService: mockAuth,
	}

	ctx := context.Background()

	// Login
	mockAuth.On("Login", ctx, &v1.LoginRequest{Email: "e", Password: "p"}).
		Return(&v1.LoginResponse{AccessToken: "a", RefreshToken: "r"}, nil)
	resp, err := client.Login(ctx, "e", "p")
	assert.NoError(t, err)
	assert.Equal(t, "a", resp.AccessToken)

	// Refresh
	mockAuth.On("Refresh", ctx, &v1.RefreshRequest{RefreshToken: "r"}).
		Return(&v1.RefreshResponse{AccessToken: "a2"}, nil)
	respR, err := client.Refresh(ctx, "r")
	assert.NoError(t, err)
	assert.Equal(t, "a2", respR.AccessToken)

	// Logout
	mockAuth.On("Logout", ctx, &v1.LogoutRequest{RefreshToken: "r"}).
		Return(&v1.LogoutResponse{Success: true}, nil)
	respL, err := client.Logout(ctx, "r")
	assert.NoError(t, err)
	assert.True(t, respL.Success)
}

func TestClient_NilServices(t *testing.T) {
	client := &agrpc.Client{}

	_, err := client.Ping(context.Background())
	assert.Error(t, err)

	_, err = client.StartScan(context.Background(), "img")
	assert.Error(t, err)

	_, err = client.GetScanStatus(context.Background(), "s1")
	assert.Error(t, err)

	_, err = client.ListScans(context.Background())
	assert.Error(t, err)

	_, err = client.GetScanReport(context.Background(), "s1")
	assert.Error(t, err)

	_, err = client.GetVulnerabilities(context.Background(), "s1")
	assert.Error(t, err)

	_, err = client.GetEvidences(context.Background(), "v1")
	assert.Error(t, err)

	_, err = client.WatchScanStatus(context.Background(), "s1")
	assert.Error(t, err)

	_, err = client.RegisterAgent(context.Background(), "t", "n")
	assert.Error(t, err)

	_, err = client.UpdateAgentStatus(context.Background(), "a", "s")
	assert.Error(t, err)

	_, err = client.GetUploadLink(context.Background(), "a", "f")
	assert.Error(t, err)
}

func TestClient_UpdateMethods(t *testing.T) {
	mockAuth := new(MockAuthServiceClient)
	client := &agrpc.Client{
		AuthService: mockAuth,
	}

	ctx := context.Background()

	// UpdateProfile
	mockAuth.On("UpdateProfile", ctx, &v1.UpdateProfileRequest{Name: "New", AvatarUrl: ""}).
		Return(&v1.UpdateProfileResponse{Success: true}, nil)
	respP, err := client.UpdateProfile(ctx, "New", "")
	assert.NoError(t, err)
	assert.True(t, respP.Success)

	// UpdateEmail
	mockAuth.On("UpdateEmail", ctx, &v1.UpdateEmailRequest{NewEmail: "new@e.com"}).
		Return(&v1.UpdateEmailResponse{Success: true}, nil)
	respE, err := client.UpdateEmail(ctx, "new@e.com")
	assert.NoError(t, err)
	assert.True(t, respE.Success)

	// UpdatePassword
	mockAuth.On("UpdatePassword", ctx, &v1.UpdatePasswordRequest{OldPassword: "o", NewPassword: "n"}).
		Return(&v1.UpdatePasswordResponse{Success: true}, nil)
	respW, err := client.UpdatePassword(ctx, "o", "n")
	assert.NoError(t, err)
	assert.True(t, respW.Success)
}

func TestClient_CompanyMethods(t *testing.T) {
	mockCompany := new(MockCompanyServiceClient)
	client := &agrpc.Client{
		CompanyService: mockCompany,
	}

	ctx := context.Background()

	// CreateCompany
	mockCompany.On("CreateCompany", ctx, &v1.CreateCompanyRequest{Name: "C", OwnerEmail: "e"}).
		Return(&v1.CreateCompanyResponse{Id: "id", Name: "C"}, nil)
	respC, err := client.CreateCompany(ctx, "C", "e")
	assert.NoError(t, err)
	assert.Equal(t, "id", respC.Id)

	// ListCompanies
	mockCompany.On("ListCompanies", ctx, &v1.ListCompaniesRequest{}).
		Return(&v1.ListCompaniesResponse{Companies: []*v1.CompanySummary{}}, nil)
	respL, err := client.ListCompanies(ctx)
	assert.NoError(t, err)
	assert.Len(t, respL, 0)
}

func TestNewClient(t *testing.T) {
	// Should succeed in creating the structure even if connection is lazy/not established yet
	c, err := agrpc.NewClient("localhost:1234", agrpc.TLSConfig{})
	assert.NoError(t, err)
	assert.NotNil(t, c)
	defer func() { _ = c.Close() }()
}

func TestClient_GetMe(t *testing.T) {
	mockAuth := new(MockAuthServiceClient)
	client := &agrpc.Client{
		AuthService: mockAuth,
	}

	ctx := context.Background()
	mockAuth.On("GetMe", mock.Anything, &v1.GetMeRequest{}).
		Return(&v1.GetMeResponse{Id: "u1"}, nil)

	resp, err := client.GetMe(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "u1", resp.Id)
}

func TestClient_TLSLoading_Error(t *testing.T) {
	// Provide invalid paths to trigger errors in loadTLSCredentials
	conf := agrpc.TLSConfig{
		Enable:   true,
		CAPath:   "nonexistent_ca",
		CertPath: "nonexistent_cert",
		KeyPath:  "nonexistent_key",
	}

	_, err := agrpc.NewClient("localhost:1234", conf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load TLS credentials")
}
func TestClient_AdminMethods(t *testing.T) {
	mockCompany := new(MockCompanyServiceClient)
	client := &agrpc.Client{
		CompanyService: mockCompany,
	}

	ctx := context.Background()

	// SearchCompanies
	mockCompany.On("ListCompanies", mock.Anything, &v1.ListCompaniesRequest{}).
		Return(&v1.ListCompaniesResponse{
			Companies: []*v1.CompanySummary{
				{Id: "c1", Name: "Company 1"},
			},
		}, nil).Once()

	respS, err := client.SearchCompanies(ctx, "test")
	assert.NoError(t, err)
	assert.Len(t, respS, 1)
	assert.Equal(t, "c1", respS[0].Id)

	// SearchUsers
	mockCompany.On("ListCompanies", mock.Anything, &v1.ListCompaniesRequest{}).
		Return(&v1.ListCompaniesResponse{
			Companies: []*v1.CompanySummary{
				{Id: "u1", Name: "User 1"},
			},
		}, nil).Once()

	respU, err := client.SearchUsers(ctx, "test", "c1")
	assert.NoError(t, err)
	assert.Len(t, respU, 1)
	assert.Equal(t, "u1", respU[0].Id)

	// AdminCreateUser
	mockCompany.On("CreateCompany", mock.Anything, &v1.CreateCompanyRequest{
		Name:       "New User",
		OwnerEmail: "user@test.com",
	}).Return(&v1.CreateCompanyResponse{Id: "u2"}, nil)

	respC, err := client.AdminCreateUser(ctx, "New User", "user@test.com", "pass1234", "admin", "c1")
	assert.NoError(t, err)
	assert.Equal(t, "u2", respC.Id)
}

func TestClient_RemoveAvatar(t *testing.T) {
	mockAuth := new(MockAuthServiceClient)
	client := &agrpc.Client{
		AuthService: mockAuth,
	}

	mockAuth.On("RemoveAvatar", mock.Anything, &v1.RemoveAvatarRequest{}).
		Return(&v1.RemoveAvatarResponse{Success: true}, nil)

	resp, err := client.RemoveAvatar(context.Background())
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestClient_OnboardCompany(t *testing.T) {
	mockCompany := new(MockCompanyServiceClient)
	client := &agrpc.Client{
		CompanyService: mockCompany,
	}

	mockCompany.On("OnboardCompany", mock.Anything, &v1.OnboardCompanyRequest{
		CompanyName:   "Co",
		OwnerName:     "Owner",
		OwnerEmail:    "e",
		OwnerPassword: "p",
	}).Return(&v1.OnboardCompanyResponse{CompanyId: "c1"}, nil)

	resp, err := client.OnboardCompany(context.Background(), "Co", "Owner", "e", "p")
	assert.NoError(t, err)
	assert.Equal(t, "c1", resp.CompanyId)
}

func TestClient_AdminMethods_Errors(t *testing.T) {
	mockCompany := new(MockCompanyServiceClient)
	client := &agrpc.Client{
		CompanyService: mockCompany,
	}
	ctx := context.Background()

	mockCompany.On("ListCompanies", mock.Anything, mock.Anything).Return((*v1.ListCompaniesResponse)(nil), fmt.Errorf("rpc error"))
	mockCompany.On("CreateCompany", mock.Anything, mock.Anything).Return((*v1.CreateCompanyResponse)(nil), fmt.Errorf("rpc error"))
	mockCompany.On("OnboardCompany", mock.Anything, mock.Anything).Return((*v1.OnboardCompanyResponse)(nil), fmt.Errorf("rpc error"))

	_, err := client.SearchCompanies(ctx, "q")
	assert.Error(t, err)

	_, err = client.SearchUsers(ctx, "q", "c1")
	assert.Error(t, err)

	_, err = client.AdminCreateUser(ctx, "n", "e", "p", "r", "c1")
	assert.Error(t, err)

	_, err = client.OnboardCompany(ctx, "c", "n", "e", "p")
	assert.Error(t, err)

	// Nil service error cases
	client.CompanyService = nil
	_, err = client.SearchCompanies(ctx, "q")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = client.SearchUsers(ctx, "q", "c1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = client.AdminCreateUser(ctx, "n", "e", "p", "r", "c1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = client.OnboardCompany(ctx, "c", "n", "e", "p")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	client.BillingService = nil
	_, err = client.GetBalance(ctx, "c1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = client.GetLedger(ctx, "c1", 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = client.GetUsageStats(ctx, "c1", 30)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	client.AuthService = nil
	_, err = client.RemoveAvatar(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestClient_BillingMethods(t *testing.T) {
	mockBilling := new(MockBillingServiceClient)
	client := &agrpc.Client{
		BillingService: mockBilling,
	}
	ctx := context.Background()

	// GetBalance
	mockBilling.On("GetBalance", mock.Anything, &v1.GetBalanceRequest{CompanyId: "c1"}).
		Return(&v1.GetBalanceResponse{Balance: 100}, nil)
	balance, err := client.GetBalance(ctx, "c1")
	assert.NoError(t, err)
	assert.Equal(t, int64(100), balance.Balance)

	// GetLedger
	mockBilling.On("GetLedger", mock.Anything, &v1.GetLedgerRequest{CompanyId: "c1", Limit: 10, Offset: 0}).
		Return(&v1.GetLedgerResponse{Entries: []*v1.LedgerEntry{}}, nil)
	ledger, err := client.GetLedger(ctx, "c1", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, ledger.Entries, 0)

	// GetUsageStats
	mockBilling.On("GetUsageStats", mock.Anything, &v1.GetUsageStatsRequest{CompanyId: "c1", Days: 30}).
		Return(&v1.GetUsageStatsResponse{Days: []*v1.UsageDay{}}, nil)
	stats, err := client.GetUsageStats(ctx, "c1", 30)
	assert.NoError(t, err)
	assert.Len(t, stats.Days, 0)

	// PreFlightCheck
	mockBilling.On("PreFlightCheck", mock.Anything, mock.Anything).
		Return(&v1.PreFlightCheckResponse{SufficientBalance: true}, nil)
	allowed, err := client.PreFlightCheck(ctx, "c1", 1, 0, 0)
	assert.NoError(t, err)
	assert.True(t, allowed.SufficientBalance)
}

func TestWithMetadata(t *testing.T) {
	ctx := context.Background()
	// Test empty context
	newCtx := agrpc.WithMetadata(ctx)
	assert.Equal(t, ctx, newCtx)

	// Test full context
	ctx = context.WithValue(ctx, types.UserIDKey, "u1")
	ctx = context.WithValue(ctx, types.CompanyIDKey, "c1")
	ctx = context.WithValue(ctx, types.RoleKey, "admin")
	ctx = context.WithValue(ctx, types.TokenKey, "secret_token")

	newCtx = agrpc.WithMetadata(ctx)
	assert.NotEqual(t, ctx, newCtx)

	md, ok := metadata.FromOutgoingContext(newCtx)
	assert.True(t, ok)
	assert.Equal(t, []string{"u1"}, md.Get("user-id"))
	assert.Equal(t, []string{"c1"}, md.Get("company-id"))
	assert.Equal(t, []string{"admin"}, md.Get("role"))
	assert.Equal(t, []string{"Bearer secret_token"}, md.Get("authorization"))
}

func TestClient_VerifyAgentSecret(t *testing.T) {
	mockAgent := new(MockAgentServiceClient)
	client := &agrpc.Client{
		AgentService: mockAgent,
	}

	ctx := context.Background()
	mockAgent.On("VerifyAgentSecret", ctx, &v1.VerifyAgentSecretRequest{AgentId: "a1", Secret: "s1"}).
		Return(&v1.VerifyAgentSecretResponse{Valid: true, TenantId: "t1"}, nil)

	resp, err := client.VerifyAgentSecret(ctx, "a1", "s1")
	assert.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.Equal(t, "t1", resp.TenantId)
}

func TestClient_VerifyToken(t *testing.T) {
	mockAuth := new(MockInternalAuthServiceClient) // I need to check if this exists
	client := &agrpc.Client{
		InternalAuthService: mockAuth,
	}

	ctx := context.Background()
	mockAuth.On("VerifyToken", ctx, &v1.VerifyTokenRequest{Token: "t1"}).
		Return(&v1.VerifyTokenResponse{Valid: true, TenantId: "tenant1"}, nil)

	resp, err := client.VerifyToken(ctx, "t1")
	assert.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.Equal(t, "tenant1", resp.TenantId)
}

func TestClient_AgentMethods(t *testing.T) {
	mockAgent := new(MockAgentServiceClient)
	client := &agrpc.Client{
		AgentService: mockAgent,
	}

	ctx := context.Background()

	// RegisterAgent
	mockAgent.On("RegisterAgent", ctx, &v1.RegisterAgentRequest{Token: "t1", Name: "n1"}).
		Return(&v1.RegisterAgentResponse{AgentId: "a1"}, nil)
	respR, err := client.RegisterAgent(ctx, "t1", "n1")
	assert.NoError(t, err)
	assert.Equal(t, "a1", respR.AgentId)

	// UpdateAgentStatus
	mockAgent.On("UpdateAgentStatus", ctx, &v1.UpdateAgentStatusRequest{AgentId: "a1", Status: "IDLE"}).
		Return(&v1.UpdateAgentStatusResponse{Success: true}, nil)
	respU, err := client.UpdateAgentStatus(ctx, "a1", "IDLE")
	assert.NoError(t, err)
	assert.True(t, respU.Success)

	// GetUploadLink
	mockAgent.On("GetUploadLink", ctx, &v1.GetUploadLinkRequest{AgentId: "a1", Filename: "f1"}).
		Return(&v1.GetUploadLinkResponse{Url: "http://minio/f1"}, nil)
	respG, err := client.GetUploadLink(ctx, "a1", "f1")
	assert.NoError(t, err)
	assert.Equal(t, "http://minio/f1", respG.Url)
}

type MockInternalAuthServiceClient struct {
	mock.Mock
}

func (m *MockInternalAuthServiceClient) VerifyToken(ctx context.Context, in *v1.VerifyTokenRequest, opts ...grpc.CallOption) (*v1.VerifyTokenResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.VerifyTokenResponse), args.Error(1)
}

func TestClient_SearchMethods(t *testing.T) {
	mockCompany := new(MockCompanyServiceClient)
	client := &agrpc.Client{CompanyService: mockCompany}
	ctx := context.Background()

	mockCompany.On("ListCompanies", mock.Anything, mock.Anything).Return(&v1.ListCompaniesResponse{}, nil)
	_, err := client.SearchCompanies(ctx, "q")
	assert.NoError(t, err)

	_, err = client.SearchUsers(ctx, "q", "c1")
	assert.NoError(t, err)
}

func TestClient_BillingPreFlight(t *testing.T) {
	mockBilling := new(MockBillingServiceClient)
	client := &agrpc.Client{BillingService: mockBilling}
	ctx := context.Background()

	mockBilling.On("PreFlightCheck", mock.Anything, mock.Anything).Return(&v1.PreFlightCheckResponse{}, nil)
	_, err := client.PreFlightCheck(ctx, "c1", 1, 1, 1)
	assert.NoError(t, err)
}

func TestClient_AdjustTokens(t *testing.T) {
	mockBilling := new(MockBillingServiceClient)
	client := &agrpc.Client{BillingService: mockBilling}
	ctx := context.Background()

	mockBilling.On("AdjustTokens", mock.Anything, mock.Anything).Return(&v1.AdjustTokensResponse{}, nil)
	_, err := client.AdjustTokens(ctx, "c1", 100, "reason")
	assert.NoError(t, err)
}

func TestClient_RemainingMethods(t *testing.T) {
	mockAuth := new(MockAuthServiceClient)
	mockBilling := new(MockBillingServiceClient)
	client := &agrpc.Client{
		AuthService: mockAuth,
		BillingService: mockBilling,
	}
	ctx := context.Background()

	mockAuth.On("GetMe", mock.Anything, mock.Anything).Return(&v1.GetMeResponse{}, nil)
	_, _ = client.GetMe(ctx)

	mockAuth.On("UpdateProfile", mock.Anything, mock.Anything).Return(&v1.UpdateProfileResponse{}, nil)
	_, _ = client.UpdateProfile(ctx, "n", "a")

	mockAuth.On("UpdateEmail", mock.Anything, mock.Anything).Return(&v1.UpdateEmailResponse{}, nil)
	_, _ = client.UpdateEmail(ctx, "e")

	mockAuth.On("UpdatePassword", mock.Anything, mock.Anything).Return(&v1.UpdatePasswordResponse{}, nil)
	_, _ = client.UpdatePassword(ctx, "o", "n")

	mockBilling.On("GetUsageStats", mock.Anything, mock.Anything).Return(&v1.GetUsageStatsResponse{}, nil)
	_, _ = client.GetUsageStats(ctx, "c", 30)
}
