package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/grpc"
	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/grpc/aegis/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MockScanServiceClient struct {
	mock.Mock
}

func (m *MockScanServiceClient) StartScan(ctx context.Context, in *v1.StartScanRequest, opts ...grpc.CallOption) (*v1.StartScanResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.StartScanResponse), args.Error(1)
}

func (m *MockScanServiceClient) GetScanStatus(ctx context.Context, in *v1.GetScanStatusRequest, opts ...grpc.CallOption) (*v1.GetScanStatusResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.GetScanStatusResponse), args.Error(1)
}

func (m *MockScanServiceClient) GetScanReport(ctx context.Context, in *v1.GetScanReportRequest, opts ...grpc.CallOption) (*v1.GetScanReportResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.GetScanReportResponse), args.Error(1)
}

func (m *MockScanServiceClient) ListScans(ctx context.Context, in *v1.ListScansRequest, opts ...grpc.CallOption) (*v1.ListScansResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.ListScansResponse), args.Error(1)
}

func TestCreateScanHandler(t *testing.T) {
	mockService := new(MockScanServiceClient)
	api := &API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}

	payload := map[string]string{"target_image": "nginx:latest"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/scans", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	mockService.On("StartScan", mock.Anything, &v1.StartScanRequest{TargetImage: "nginx:latest"}).
		Return(&v1.StartScanResponse{ScanId: "s1"}, nil)

	handler := http.HandlerFunc(api.CreateScanHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestCreateScanHandler_GRPCFailure(t *testing.T) {
	mockService := new(MockScanServiceClient)
	api := &API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}

	payload := map[string]string{"target_image": "nginx:latest"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/scans", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	mockService.On("StartScan", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("grpc error"))

	handler := http.HandlerFunc(api.CreateScanHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestGetScansHandler(t *testing.T) {
	mockService := new(MockScanServiceClient)
	api := &API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}

	resp := &v1.ListScansResponse{
		Scans: []*v1.ScanDetails{
			{
				ScanId:             "s1",
				TemporalWorkflowId: "wf-1",
				TargetImage:        "img-1",
				Status:             "PENDING",
				StartedAt:          timestamppb.New(time.Now()),
				CompletedAt:        nil,
			},
		},
	}

	mockService.On("ListScans", mock.Anything, &v1.ListScansRequest{}).
		Return(resp, nil)

	req, _ := http.NewRequest("GET", "/scans", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetScansHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetScansHandler_GRPCError(t *testing.T) {
	mockService := new(MockScanServiceClient)
	api := &API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}

	mockService.On("ListScans", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("grpc error"))

	req, _ := http.NewRequest("GET", "/scans", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetScansHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestGetScanByIDHandler_Found(t *testing.T) {
	mockService := new(MockScanServiceClient)
	api := &API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}

	resp := &v1.ListScansResponse{
		Scans: []*v1.ScanDetails{
			{
				ScanId:             "s1",
				TemporalWorkflowId: "wf-1",
				TargetImage:        "img-1",
				Status:             "PENDING",
				StartedAt:          timestamppb.New(time.Now()),
				CompletedAt:        nil,
			},
		},
	}

	mockService.On("ListScans", mock.Anything, &v1.ListScansRequest{}).
		Return(resp, nil)

	req, _ := http.NewRequest("GET", "/scans/s1", nil)
	req.SetPathValue("id", "s1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetScanByIDHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetScanByIDHandler_NotFound(t *testing.T) {
	mockService := new(MockScanServiceClient)
	api := &API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}

	resp := &v1.ListScansResponse{
		Scans: []*v1.ScanDetails{},
	}

	mockService.On("ListScans", mock.Anything, &v1.ListScansRequest{}).
		Return(resp, nil)

	req, _ := http.NewRequest("GET", "/scans/nb", nil)
	req.SetPathValue("id", "nb")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetScanByIDHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetScanReportHandler_Success(t *testing.T) {
	mockService := new(MockScanServiceClient)
	api := &API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}

	resp := &v1.GetScanReportResponse{
		PdfData: []byte("fake-pdf-content"),
	}

	mockService.On("GetScanReport", mock.Anything, &v1.GetScanReportRequest{ScanId: "s1"}).
		Return(resp, nil)

	req, _ := http.NewRequest("GET", "/scans/s1/report", nil)
	req.SetPathValue("id", "s1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetScanReportHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/pdf", rr.Header().Get("Content-Type"))
}

func TestGetScanReportHandler_EmptyReport(t *testing.T) {
	mockService := new(MockScanServiceClient)
	api := &API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}

	resp := &v1.GetScanReportResponse{
		PdfData: []byte{},
	}

	mockService.On("GetScanReport", mock.Anything, &v1.GetScanReportRequest{ScanId: "s1"}).
		Return(resp, nil)

	req, _ := http.NewRequest("GET", "/scans/s1/report", nil)
	req.SetPathValue("id", "s1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetScanReportHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetScanReportHandler_GRPCError(t *testing.T) {
	mockService := new(MockScanServiceClient)
	api := &API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}

	mockService.On("GetScanReport", mock.Anything, &v1.GetScanReportRequest{ScanId: "s1"}).
		Return(nil, fmt.Errorf("grpc error"))

	req, _ := http.NewRequest("GET", "/scans/s1/report", nil)
	req.SetPathValue("id", "s1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetScanReportHandler)
	handler.ServeHTTP(rr, req)

	// Since we return 404 for GRPC errors in GetScanReportHandler for simplicity
	assert.Equal(t, http.StatusNotFound, rr.Code)
}
