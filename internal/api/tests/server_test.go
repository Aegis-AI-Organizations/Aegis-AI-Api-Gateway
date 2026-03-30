package api_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api"
	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/grpc"
	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/grpc/aegis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

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

type closeNotifierRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func (c *closeNotifierRecorder) CloseNotify() <-chan bool {
	return c.closed
}

func newCloseNotifierRecorder() *closeNotifierRecorder {
	return &closeNotifierRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		closed:           make(chan bool, 1),
	}
}

func TestNewRouterFull(t *testing.T) {
	dummyClient := &agrpc.Client{
		ScanService:          &MockScanServiceClient{},
		VulnerabilityService: &MockVulnerabilityServiceClient{},
		AuthService:          &MockAuthServiceClient{},
	}
	mux := api.NewRouter(dummyClient)

	tests := []struct {
		method string
		path   string
		code   int
	}{
		{"GET", "/health", http.StatusOK},
		{"GET", "/", http.StatusOK},
		{"POST", "/scans", http.StatusCreated}, // Mock returns success
		{"GET", "/scans", http.StatusOK},
		{"GET", "/scans/1", http.StatusOK},
		{"GET", "/scans/1/vulnerabilities", http.StatusOK},
		{"GET", "/vulnerabilities/1/evidences", http.StatusOK},
		{"GET", "/scans/1/report", http.StatusOK},
		{"GET", "/scans/stream", http.StatusOK},
		{"GET", "/scans/1/stream", http.StatusOK},
	}

	for _, tt := range tests {
		var body []byte
		if tt.method == "POST" {
			body = []byte(`{"target_image":"test"}`)
		}
		req, _ := http.NewRequest(tt.method, tt.path, bytes.NewBuffer(body))
		rr := newCloseNotifierRecorder()
		mux.ServeHTTP(rr, req)
		assert.NotEqual(t, http.StatusNotFound, rr.Code, "Path %s %s should be registered", tt.method, tt.path)
		assert.Equal(t, tt.code, rr.Code, "Path %s %s should return code %d", tt.method, tt.path, tt.code)
	}
}
