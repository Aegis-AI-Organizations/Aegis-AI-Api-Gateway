package api_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api"
	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/grpc"
	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/grpc/aegis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
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
func (m *MockScanServiceClient) WatchScanStatus(ctx context.Context, in *v1.WatchScanStatusRequest, opts ...grpc.CallOption) (v1.ScanService_WatchScanStatusClient, error) {
	return nil, nil
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

func TestNewRouterFull(t *testing.T) {
	dummyClient := &agrpc.Client{
		ScanService:          &MockScanServiceClient{},
		VulnerabilityService: &MockVulnerabilityServiceClient{},
	}
	mux := api.NewRouter(dummyClient)

	tests := []struct {
		method string
		path   string
		code   int
	}{
		{"GET", "/health", http.StatusOK},
		{"GET", "/", http.StatusOK},
		{"POST", "/scans", http.StatusBadRequest},
		{"GET", "/scans", http.StatusInternalServerError},
		{"GET", "/scans/1", http.StatusInternalServerError},
		{"GET", "/scans/1/vulnerabilities", http.StatusInternalServerError},
		{"GET", "/vulnerabilities/1/evidences", http.StatusInternalServerError},
		{"GET", "/scans/1/report", http.StatusInternalServerError},
		{"GET", "/scans/stream", http.StatusOK},
		{"GET", "/scans/1/stream", http.StatusOK},
	}

	for _, tt := range tests {
		var body []byte
		if tt.method == "POST" {
			body = []byte(`{"target_image":"test"}`)
		}
		req, _ := http.NewRequest(tt.method, tt.path, bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		assert.NotEqual(t, http.StatusNotFound, rr.Code, "Path %s %s should be registered", tt.method, tt.path)
	}
}
