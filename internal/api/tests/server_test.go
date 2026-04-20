package api_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/testutils"
	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc/aegis/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Helper to generate a test token
func generateTestToken(role, secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":        "test-user",
		"company_id": "test-company",
		"role":       role,
	})
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

type MockScanServiceClient struct {
	mock.Mock
}

func (m *MockScanServiceClient) StartScan(ctx context.Context, in *v1.StartScanRequest, opts ...grpc.CallOption) (*v1.StartScanResponse, error) {
	return &v1.StartScanResponse{ScanId: "test-id"}, nil
}
func (m *MockScanServiceClient) GetScanStatus(ctx context.Context, in *v1.GetScanStatusRequest, opts ...grpc.CallOption) (*v1.GetScanStatusResponse, error) {
	return &v1.GetScanStatusResponse{}, nil
}
func (m *MockScanServiceClient) GetScanReport(ctx context.Context, in *v1.GetScanReportRequest, opts ...grpc.CallOption) (*v1.GetScanReportResponse, error) {
	return &v1.GetScanReportResponse{PdfData: []byte("pdf")}, nil
}
func (m *MockScanServiceClient) ListScans(ctx context.Context, in *v1.ListScansRequest, opts ...grpc.CallOption) (*v1.ListScansResponse, error) {
	return &v1.ListScansResponse{
		Scans: []*v1.ScanDetails{{ScanId: "1"}},
	}, nil
}
type MockScanStreamClient struct {
	mock.Mock
}

func (m *MockScanStreamClient) Recv() (*v1.WatchScanStatusResponse, error) {
	return nil, fmt.Errorf("EOF")
}

func (m *MockScanStreamClient) Context() context.Context { return context.Background() }
func (m *MockScanStreamClient) Header() (metadata.MD, error) { return nil, nil }
func (m *MockScanStreamClient) Trailer() metadata.MD { return nil }
func (m *MockScanStreamClient) CloseSend() error { return nil }
func (m *MockScanStreamClient) SendMsg(m_ interface{}) error { return nil }
func (m *MockScanStreamClient) RecvMsg(m_ interface{}) error { return nil }

func (m *MockScanServiceClient) WatchScanStatus(ctx context.Context, in *v1.WatchScanStatusRequest, opts ...grpc.CallOption) (v1.ScanService_WatchScanStatusClient, error) {
	return &MockScanStreamClient{}, nil
}

type MockVulnerabilityServiceClient struct {
	mock.Mock
}
func (m *MockVulnerabilityServiceClient) GetVulnerabilities(ctx context.Context, in *v1.GetVulnerabilitiesRequest, opts ...grpc.CallOption) (*v1.GetVulnerabilitiesResponse, error) {
	return &v1.GetVulnerabilitiesResponse{}, nil
}
func (m *MockVulnerabilityServiceClient) GetEvidences(ctx context.Context, in *v1.GetEvidencesRequest, opts ...grpc.CallOption) (*v1.GetEvidencesResponse, error) {
	return &v1.GetEvidencesResponse{}, nil
}

type MockAuthServiceClient struct {
	mock.Mock
}

func (m *MockAuthServiceClient) Login(ctx context.Context, in *v1.LoginRequest, opts ...grpc.CallOption) (*v1.LoginResponse, error) {
	return &v1.LoginResponse{AccessToken: "at", RefreshToken: "rt"}, nil
}
func (m *MockAuthServiceClient) Refresh(ctx context.Context, in *v1.RefreshRequest, opts ...grpc.CallOption) (*v1.RefreshResponse, error) {
	return &v1.RefreshResponse{AccessToken: "at"}, nil
}
func (m *MockAuthServiceClient) Logout(ctx context.Context, in *v1.LogoutRequest, opts ...grpc.CallOption) (*v1.LogoutResponse, error) {
	return &v1.LogoutResponse{Success: true}, nil
}
func (m *MockAuthServiceClient) GetMe(ctx context.Context, in *v1.GetMeRequest, opts ...grpc.CallOption) (*v1.GetMeResponse, error) {
	return &v1.GetMeResponse{}, nil
}
func (m *MockAuthServiceClient) UpdateProfile(ctx context.Context, in *v1.UpdateProfileRequest, opts ...grpc.CallOption) (*v1.UpdateProfileResponse, error) {
	return &v1.UpdateProfileResponse{Success: true}, nil
}
func (m *MockAuthServiceClient) UpdateEmail(ctx context.Context, in *v1.UpdateEmailRequest, opts ...grpc.CallOption) (*v1.UpdateEmailResponse, error) {
	return &v1.UpdateEmailResponse{Success: true}, nil
}
func (m *MockAuthServiceClient) UpdatePassword(ctx context.Context, in *v1.UpdatePasswordRequest, opts ...grpc.CallOption) (*v1.UpdatePasswordResponse, error) {
	return &v1.UpdatePasswordResponse{Success: true}, nil
}

type MockCompanyServiceClient struct {
	mock.Mock
}
func (m *MockCompanyServiceClient) CreateCompany(ctx context.Context, in *v1.CreateCompanyRequest, opts ...grpc.CallOption) (*v1.CreateCompanyResponse, error) {
	return &v1.CreateCompanyResponse{Id: "1", Name: in.Name}, nil
}
func (m *MockCompanyServiceClient) ListCompanies(ctx context.Context, in *v1.ListCompaniesRequest, opts ...grpc.CallOption) (*v1.ListCompaniesResponse, error) {
	return &v1.ListCompaniesResponse{Companies: []*v1.CompanySummary{{Id: "1", Name: "Test"}}}, nil
}


func TestNewRouterFull(t *testing.T) {
	testSecret := "test-secret-123"
	t.Setenv("JWT_SECRET", testSecret)
	token := generateTestToken("superadmin", testSecret)

	dummyClient := &agrpc.Client{
		ScanService:          &MockScanServiceClient{},
		VulnerabilityService: &MockVulnerabilityServiceClient{},
		AuthService:          &MockAuthServiceClient{},
		CompanyService:       &MockCompanyServiceClient{},
	}
	mux := api.NewRouter(dummyClient)

	tests := []struct {
		method    string
		path      string
		code      int
		protected bool
	}{
		{"GET", "/health", http.StatusOK, false},
		{"GET", "/", http.StatusOK, false},
		{"POST", "/scans", http.StatusCreated, true},
		{"GET", "/scans", http.StatusOK, true},
		{"GET", "/scans/1", http.StatusOK, true},
		{"GET", "/scans/1/vulnerabilities", http.StatusOK, true},
		{"GET", "/vulnerabilities/1/evidences", http.StatusOK, true},
		{"GET", "/scans/1/report", http.StatusOK, true},
		{"GET", "/scans/stream", http.StatusOK, true},
		{"GET", "/scans/1/stream", http.StatusOK, true},
	}

	for _, tt := range tests {
		var body []byte
		if tt.method == "POST" {
			body = []byte(`{"target_image":"test"}`)
		}
		req, _ := http.NewRequest(tt.method, tt.path, bytes.NewBuffer(body))
		if tt.protected {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		rr := testutils.NewCloseNotifierRecorder()
		mux.ServeHTTP(rr, req)
		assert.NotEqual(t, http.StatusNotFound, rr.Code, "Path %s %s should be registered", tt.method, tt.path)
		assert.Equal(t, tt.code, rr.Code, "Path %s %s should return code %d", tt.method, tt.path, tt.code)
	}
}

func TestScopesForbidden(t *testing.T) {
	testSecret := "test-secret-123"
	t.Setenv("JWT_SECRET", testSecret)

	dummyClient := &agrpc.Client{
		ScanService:          &MockScanServiceClient{},
		VulnerabilityService: &MockVulnerabilityServiceClient{},
		AuthService:          &MockAuthServiceClient{},
		CompanyService:       &MockCompanyServiceClient{},
	}
	mux := api.NewRouter(dummyClient)

	// A viewer should NOT be able to post a scan (requires scan:write)
	viewerToken := generateTestToken("viewer", testSecret)

	tests := []struct {
		name   string
		method string
		path   string
		token  string
		want   int
	}{
		{
			name:   "Viewer cannot launch scan",
			method: "POST",
			path:   "/scans",
			token:  viewerToken,
			want:   http.StatusForbidden,
		},
		{
			name:   "Operator can launch scan",
			method: "POST",
			path:   "/scans",
			token:  generateTestToken("operator", testSecret),
			want:   http.StatusCreated,
		},
		{
			name:   "Viewer cannot create company",
			method: "POST",
			path:   "/companies",
			token:  viewerToken,
			want:   http.StatusForbidden,
		},
		{
			name:   "SuperAdmin can create company",
			method: "POST",
			path:   "/companies",
			token:  generateTestToken("superadmin", testSecret),
			want:   http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.method == "POST" {
				body = []byte(`{"target_image":"test", "name":"test", "owner_email":"test@test.com"}`)
			}
			req, _ := http.NewRequest(tt.method, tt.path, bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+tt.token)

			rr := testutils.NewCloseNotifierRecorder()
			mux.ServeHTTP(rr, req)

			assert.Equal(t, tt.want, rr.Code)
		})
	}
}
