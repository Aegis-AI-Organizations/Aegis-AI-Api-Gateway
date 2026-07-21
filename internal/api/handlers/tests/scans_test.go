package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc/aegis/v2"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (m *MockScanServiceClient) WatchScanStatus(ctx context.Context, in *v1.WatchScanStatusRequest, opts ...grpc.CallOption) (v1.ScanService_WatchScanStatusClient, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(v1.ScanService_WatchScanStatusClient), args.Error(1)
}

func (m *MockScanServiceClient) UpdateScanStatus(ctx context.Context, in *v1.UpdateScanStatusRequest, opts ...grpc.CallOption) (*v1.UpdateScanStatusResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.UpdateScanStatusResponse), args.Error(1)
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

func (m *MockBillingServiceClient) GetLedger(ctx context.Context, in *v1.GetLedgerRequest, opts ...grpc.CallOption) (*v1.GetLedgerResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.GetLedgerResponse), args.Error(1)
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

func (m *MockBillingServiceClient) GetUsageStats(ctx context.Context, in *v1.GetUsageStatsRequest, opts ...grpc.CallOption) (*v1.GetUsageStatsResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.GetUsageStatsResponse), args.Error(1)
}

func TestCreateScanHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockScan := new(MockScanServiceClient)
	mockBilling := new(MockBillingServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			ScanService:    mockScan,
			BillingService: mockBilling,
		},
	}

	payload := map[string]string{"target_image": "nginx:latest"}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("company_id", "test-company") // Inject identity
	c.Request, _ = http.NewRequest("POST", "/scans", bytes.NewBuffer(body))

	// Mock Billing Pre-flight (Sufficient balance)
	mockBilling.On("PreFlightCheck", mock.Anything, mock.Anything).
		Return(&v1.PreFlightCheckResponse{SufficientBalance: true, EstimatedCost: 10}, nil)

	// Mock Token Deduction
	mockBilling.On("AdjustTokens", mock.Anything, mock.MatchedBy(func(req *v1.AdjustTokensRequest) bool {
		return req.CompanyId == "test-company" &&
			req.Amount == -10 &&
			req.Reason == "Scan consumption: nginx:latest"
	})).
		Return(&v1.AdjustTokensResponse{Balance: 90}, nil)

	// Mock Scan Launch
	mockScan.On("StartScan", mock.Anything, &v1.StartScanRequest{TargetImage: "nginx:latest"}).
		Return(&v1.StartScanResponse{ScanId: "s1", Status: "PENDING"}, nil)

	api.CreateScanHandler(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockScan.AssertExpectations(t)
	mockBilling.AssertExpectations(t)
}

func TestCreateScanHandler_TopologySelection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockScan := new(MockScanServiceClient)
	mockBilling := new(MockBillingServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			ScanService:    mockScan,
			BillingService: mockBilling,
		},
	}

	payload := map[string]any{
		"scope":           "topology",
		"target_node_ids": []string{"container-b", "container-a"},
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("company_id", "test-company")
	c.Request, _ = http.NewRequest("POST", "/scans", bytes.NewBuffer(body))

	mockBilling.On("PreFlightCheck", mock.Anything, mock.MatchedBy(func(req *v1.PreFlightCheckRequest) bool {
		return req.CompanyId == "test-company" &&
			req.TargetConfig != nil &&
			req.TargetConfig.IpCount == 0 &&
			req.TargetConfig.ApiCount == 0 &&
			req.TargetConfig.WebappCount == 2
	})).
		Return(&v1.PreFlightCheckResponse{SufficientBalance: true, EstimatedCost: 10}, nil)
	mockBilling.On("AdjustTokens", mock.Anything, mock.MatchedBy(func(req *v1.AdjustTokensRequest) bool {
		return req.CompanyId == "test-company" &&
			req.Amount == -10 &&
			req.Reason == "Scan consumption: topology selection (2 targets)"
	})).
		Return(&v1.AdjustTokensResponse{Balance: 90}, nil)
	mockScan.On("StartScan", mock.Anything, &v1.StartScanRequest{TargetImage: "topology:container-a,container-b"}).
		Return(&v1.StartScanResponse{ScanId: "s1", Status: "PENDING"}, nil)

	api.CreateScanHandler(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockScan.AssertExpectations(t)
	mockBilling.AssertExpectations(t)
}

func TestCreateScanHandler_TokenConsumptionPermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockScan := new(MockScanServiceClient)
	mockBilling := new(MockBillingServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			ScanService:    mockScan,
			BillingService: mockBilling,
		},
	}

	payload := map[string]string{"target_image": "nginx:latest"}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("company_id", "test-company")
	c.Request, _ = http.NewRequest("POST", "/scans", bytes.NewBuffer(body))

	mockBilling.On("PreFlightCheck", mock.Anything, mock.Anything).
		Return(&v1.PreFlightCheckResponse{SufficientBalance: true, EstimatedCost: 10}, nil)
	mockBilling.On("AdjustTokens", mock.Anything, mock.Anything).
		Return(nil, status.Error(codes.PermissionDenied, "permission denied"))

	api.CreateScanHandler(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Token consumption is not allowed")
	mockScan.AssertNotCalled(t, "StartScan", mock.Anything, mock.Anything)
	mockBilling.AssertExpectations(t)
}

func TestCreateScanHandler_GRPCFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockScan := new(MockScanServiceClient)
	mockBilling := new(MockBillingServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			ScanService:    mockScan,
			BillingService: mockBilling,
		},
	}

	payload := map[string]string{"target_image": "nginx:latest"}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("company_id", "test-company")
	c.Request, _ = http.NewRequest("POST", "/scans", bytes.NewBuffer(body))

	mockBilling.On("PreFlightCheck", mock.Anything, mock.Anything).
		Return(&v1.PreFlightCheckResponse{SufficientBalance: true, EstimatedCost: 10}, nil)
	mockBilling.On("AdjustTokens", mock.Anything, mock.Anything).
		Return(&v1.AdjustTokensResponse{Balance: 90}, nil)

	mockScan.On("StartScan", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("grpc error"))

	api.CreateScanHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateScanHandler_EmptyTargetImage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &handlers.API{}
	payload := map[string]string{"target_image": ""}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/scans", bytes.NewBuffer(body))
	api.CreateScanHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateScanHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &handlers.API{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/scans", bytes.NewBuffer([]byte("not-json")))
	api.CreateScanHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetScansHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockScanServiceClient)
	api := &handlers.API{
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
				StartedAt:          timestamppb.Now(),
				CompletedAt:        nil,
				DebugBundle:        "s3://aegis-debug/debug-bundles/s1/bundle.tar.gz",
			},
		},
	}

	mockService.On("ListScans", mock.Anything, &v1.ListScansRequest{}).
		Return(resp, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/scans", nil)

	api.GetScansHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "debug_bundle")
	assert.Contains(t, w.Body.String(), "s3://aegis-debug/debug-bundles/s1/bundle.tar.gz")
}

func TestGetScansHandler_GRPCError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockScanServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}

	mockService.On("ListScans", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("grpc error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/scans", nil)

	api.GetScansHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetScansHandler_NilRes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockScanServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}
	mockService.On("ListScans", mock.Anything, mock.Anything).Return(&v1.ListScansResponse{Scans: nil}, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/scans", nil)
	api.GetScansHandler(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetScanByIDHandler_Found(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockScanServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}

	resp := &v1.GetScanStatusResponse{
		ScanId:             "s1",
		TemporalWorkflowId: "wf-1",
		TargetImage:        "img-1",
		Status:             "PENDING",
		StartedAt:          timestamppb.Now(),
		CompletedAt:        nil,
		DebugBundle:        "s3://aegis-debug/debug-bundles/s1/bundle.tar.gz",
	}

	mockService.On("GetScanStatus", mock.Anything, &v1.GetScanStatusRequest{ScanId: "s1"}).
		Return(resp, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/scans/s1", nil)
	c.Params = []gin.Param{{Key: "id", Value: "s1"}}

	api.GetScanByIDHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "debug_bundle")
	assert.Contains(t, w.Body.String(), "s3://aegis-debug/debug-bundles/s1/bundle.tar.gz")
}

func TestGetScanByIDHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockScanServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}

	mockService.On("GetScanStatus", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("not found"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/scans/nb", nil)
	c.Params = []gin.Param{{Key: "id", Value: "nb"}}

	api.GetScanByIDHandler(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetScanByIDHandler_EmptyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &handlers.API{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/scans/", nil)
	api.GetScanByIDHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetScanReportHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockScanServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}

	resp := &v1.GetScanReportResponse{
		PdfData: []byte("fake-pdf-content"),
	}

	mockService.On("GetScanReport", mock.Anything, &v1.GetScanReportRequest{ScanId: "s1"}).
		Return(resp, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/scans/s1/report", nil)
	c.Params = []gin.Param{{Key: "id", Value: "s1"}}

	api.GetScanReportHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
}

func TestGetScanReportHandler_EmptyReport(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockScanServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}

	resp := &v1.GetScanReportResponse{
		PdfData: []byte{},
	}

	mockService.On("GetScanReport", mock.Anything, &v1.GetScanReportRequest{ScanId: "s1"}).
		Return(resp, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/scans/s1/report", nil)
	c.Params = []gin.Param{{Key: "id", Value: "s1"}}

	api.GetScanReportHandler(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetScanReportHandler_GRPCError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockScanServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}

	mockService.On("GetScanReport", mock.Anything, &v1.GetScanReportRequest{ScanId: "s1"}).
		Return(nil, fmt.Errorf("grpc error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/scans/s1/report", nil)
	c.Params = []gin.Param{{Key: "id", Value: "s1"}}

	api.GetScanReportHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
