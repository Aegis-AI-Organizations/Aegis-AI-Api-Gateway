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
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

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

func (m *MockCompanyServiceClient) RotateAgentToken(ctx context.Context, in *v1.RotateAgentTokenRequest, opts ...grpc.CallOption) (*v1.RotateAgentTokenResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.RotateAgentTokenResponse), args.Error(1)
}

func (m *MockCompanyServiceClient) RevokeAgentToken(ctx context.Context, in *v1.RevokeAgentTokenRequest, opts ...grpc.CallOption) (*v1.RevokeAgentTokenResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.RevokeAgentTokenResponse), args.Error(1)
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

type MockCompanyUpdateStream struct {
	mock.Mock
}

func (m *MockCompanyUpdateStream) Recv() (*v1.WatchCompanyUpdatesResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.WatchCompanyUpdatesResponse), args.Error(1)
}

func (m *MockCompanyUpdateStream) Context() context.Context     { return context.Background() }
func (m *MockCompanyUpdateStream) Header() (metadata.MD, error) { return nil, nil }
func (m *MockCompanyUpdateStream) Trailer() metadata.MD         { return nil }
func (m *MockCompanyUpdateStream) CloseSend() error             { return nil }
func (m *MockCompanyUpdateStream) SendMsg(m_ interface{}) error { return nil }
func (m *MockCompanyUpdateStream) RecvMsg(m_ interface{}) error { return nil }

func TestListCompaniesHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/companies", nil)

	mockCompany.On("ListCompanies", mock.Anything, mock.Anything).
		Return(&v1.ListCompaniesResponse{
			Companies: []*v1.CompanySummary{
				{Id: "c1", Name: "Company 1"},
			},
		}, nil)

	api.ListCompaniesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListCompaniesHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/companies", nil)

	mockCompany.On("ListCompanies", mock.Anything, mock.Anything).
		Return(nil, status.Error(codes.Internal, "internal error"))

	api.ListCompaniesHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateCompanyHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	payload := map[string]string{
		"name":        "New Co",
		"owner_email": "owner@example.com",
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/companies", bytes.NewBuffer(body))

	mockCompany.On("CreateCompany", mock.Anything, &v1.CreateCompanyRequest{
		Name:       "New Co",
		OwnerEmail: "owner@example.com",
	}).Return(&v1.CreateCompanyResponse{Id: "c2", Name: "New Co"}, nil)

	api.CreateCompanyHandler(c)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateCompanyHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	payload := map[string]string{
		"name":        "Fail Co",
		"owner_email": "fail@example.com",
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/companies", bytes.NewBuffer(body))

	mockCompany.On("CreateCompany", mock.Anything, mock.Anything).
		Return(nil, status.Error(codes.Internal, "internal error"))

	api.CreateCompanyHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateCompanyHandler_GRPCError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	payload := map[string]string{
		"name":        "Fail Co",
		"owner_email": "fail@example.com",
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/companies", bytes.NewBuffer(body))

	mockCompany.On("CreateCompany", mock.Anything, mock.Anything).
		Return(nil, status.Error(codes.Internal, "db error"))

	api.CreateCompanyHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
func TestOnboardCompanyHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	payload := map[string]string{
		"company_name": "Onboard Co",
		"owner_name":   "John Doe",
		"owner_email":  "john@example.com",
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/companies/onboard", bytes.NewBuffer(body))

	mockCompany.On("OnboardCompany", mock.Anything, &v1.OnboardCompanyRequest{
		CompanyName: "Onboard Co",
		OwnerName:   "John Doe",
		OwnerEmail:  "john@example.com",
	}).Return(&v1.OnboardCompanyResponse{
		CompanyId:       "c3",
		OwnerId:         "u1",
		DeploymentToken: "ag_0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefg",
	}, nil)

	api.OnboardCompanyHandler(c)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestOnboardCompanyHandler_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &handlers.API{}

	payload := map[string]string{
		"company_name": "Missing Fields",
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/companies/onboard", bytes.NewBuffer(body))

	api.OnboardCompanyHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
func TestOnboardCompanyHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	payload := map[string]string{
		"company_name": "Onboard Co",
		"owner_name":   "John Doe",
		"owner_email":  "john@example.com",
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/companies/onboard", bytes.NewBuffer(body))

	mockCompany.On("OnboardCompany", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("grpc error"))

	api.OnboardCompanyHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRotateAgentTokenHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/companies/me/agent-token/rotate", nil)

	mockCompany.On("RotateAgentToken", mock.Anything, &v1.RotateAgentTokenRequest{}).
		Return(&v1.RotateAgentTokenResponse{AgentToken: "ag_rotated-token"}, nil)

	api.RotateAgentTokenHandler(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "ag_rotated-token")
}

func TestRotateAgentTokenHandler_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/companies/me/agent-token/rotate", nil)

	mockCompany.On("RotateAgentToken", mock.Anything, &v1.RotateAgentTokenRequest{}).
		Return(nil, status.Error(codes.PermissionDenied, "forbidden"))

	api.RotateAgentTokenHandler(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminRotateAgentTokenHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "company-123"}}
	c.Request, _ = http.NewRequest("POST", "/admin/companies/company-123/agent-token/rotate", nil)

	mockCompany.On("RotateAgentToken", mock.Anything, &v1.RotateAgentTokenRequest{CompanyId: "company-123"}).
		Return(&v1.RotateAgentTokenResponse{AgentToken: "ag_admin-rotated-token"}, nil)

	api.AdminRotateAgentTokenHandler(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "ag_admin-rotated-token")
}

func TestAdminRotateAgentTokenHandler_MissingCompanyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &handlers.API{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/admin/companies//agent-token/rotate", nil)

	api.AdminRotateAgentTokenHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "company id is required")
}

func TestAdminRevokeAgentTokenHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "company-123"}}
	c.Request, _ = http.NewRequest("POST", "/admin/companies/company-123/agent-token/revoke", nil)

	mockCompany.On("RevokeAgentToken", mock.Anything, &v1.RevokeAgentTokenRequest{CompanyId: "company-123"}).
		Return(&v1.RevokeAgentTokenResponse{Success: true}, nil)

	api.AdminRevokeAgentTokenHandler(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestAdminRevokeAgentTokenHandler_MissingCompanyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &handlers.API{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/admin/companies//agent-token/revoke", nil)

	api.AdminRevokeAgentTokenHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "company id is required")
}

func TestRevokeAgentTokenHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/companies/me/agent-token/revoke", nil)

	mockCompany.On("RevokeAgentToken", mock.Anything, &v1.RevokeAgentTokenRequest{}).
		Return(&v1.RevokeAgentTokenResponse{Success: true}, nil)

	api.RevokeAgentTokenHandler(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRevokeAgentTokenHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/companies/me/agent-token/revoke", nil)

	mockCompany.On("RevokeAgentToken", mock.Anything, &v1.RevokeAgentTokenRequest{}).
		Return(nil, status.Error(codes.NotFound, "not found"))

	api.RevokeAgentTokenHandler(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
