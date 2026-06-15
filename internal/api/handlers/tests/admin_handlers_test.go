package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc/aegis/v2"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/testutils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockAdminGRPCClient struct {
	mock.Mock
}

func (m *MockAdminGRPCClient) SearchCompanies(ctx context.Context, query string) ([]*v1.CompanySummary, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*v1.CompanySummary), args.Error(1)
}

func (m *MockAdminGRPCClient) SearchUsers(ctx context.Context, query, companyID string) ([]*v1.CompanySummary, error) {
	args := m.Called(ctx, query, companyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*v1.CompanySummary), args.Error(1)
}

func (m *MockAdminGRPCClient) AdminCreateUser(ctx context.Context, name, email, password, role, companyID string) (*v1.CreateCompanyResponse, error) {
	args := m.Called(ctx, name, email, password, role, companyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.CreateCompanyResponse), args.Error(1)
}

func (m *MockAdminGRPCClient) ListAuditLogs(ctx context.Context, limit, offset int32, companyID string) (*v1.ListAuditLogsResponse, error) {
	args := m.Called(ctx, limit, offset, companyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.ListAuditLogsResponse), args.Error(1)
}

// NOTE: Since handlers.API uses agrpc.Client which is a struct with interfaces,
// we might need to mock the interfaces inside agrpc.Client or mock agrpc.Client methods if they were interfaces.
// But agrpc.Client methods are not interfaces.
// However, in handlers_test, they often just inject a dummy agrpc.Client with mocked services.
// Let's use the same approach as company_test.go.

func TestSearchCompaniesHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/admin/companies?search=test", nil)

	mockCompany.On("ListCompanies", mock.Anything, mock.Anything).
		Return(&v1.ListCompaniesResponse{
			Companies: []*v1.CompanySummary{
				{Id: "c1", Name: "Company 1", DeploymentToken: "t1", OwnerEmail: "o1"},
			},
		}, nil)

	api.SearchCompaniesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []handlers.CompanySearchResult
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, "c1", resp[0].ID)
}

func TestSearchCompaniesHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/admin/companies", nil)

	mockCompany.On("ListCompanies", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("grpc error"))

	api.SearchCompaniesHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSearchUsersHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/admin/users?search=test&company_id=c1", nil)

	mockCompany.On("ListCompanies", mock.Anything, mock.Anything).
		Return(&v1.ListCompaniesResponse{
			Companies: []*v1.CompanySummary{
				{Id: "u1", Name: "User 1", OwnerEmail: "u@t.com", DeploymentToken: "admin", OwnerId: "c1"},
			},
		}, nil)

	api.SearchUsersHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []handlers.UserSearchResult
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, "u1", resp[0].ID)
	assert.Equal(t, "c1", resp[0].CompanyID)
}

func TestSearchUsersHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/admin/users", nil)

	mockCompany.On("ListCompanies", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("grpc error"))

	api.SearchUsersHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateUserHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	payload := map[string]string{
		"name":       "New User",
		"email":      "user@test.com",
		"password":   "password123",
		"role":       "admin",
		"company_id": "c1",
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/admin/users", bytes.NewBuffer(body))

	mockCompany.On("CreateCompany", mock.Anything, mock.Anything).
		Return(&v1.CreateCompanyResponse{Id: "u2"}, nil)

	api.CreateUserHandler(c)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateUserHandler_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &handlers.API{}

	payload := map[string]string{
		"name": "Missing Fields",
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/admin/users", bytes.NewBuffer(body))

	api.CreateUserHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateUserHandler_ServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	payload := map[string]string{
		"name":       "New User",
		"email":      "user@test.com",
		"password":   "password123",
		"role":       "admin",
		"company_id": "c1",
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/admin/users", bytes.NewBuffer(body))

	mockCompany.On("CreateCompany", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("grpc error"))

	api.CreateUserHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListTenantUsersHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("company_id", "tenant-1")
	c.Request, _ = http.NewRequest("GET", "/users?search=john", nil)

	mockCompany.On("ListCompanies", mock.Anything, mock.Anything).
		Return(&v1.ListCompaniesResponse{
			Companies: []*v1.CompanySummary{
				{
					Id:              "u1",
					Name:            "John Doe",
					OwnerEmail:      "john@test.com",
					DeploymentToken: "viewer",
					OwnerId:         "tenant-1",
					OrgType:         v1.OrganizationType_ORGANIZATION_TYPE_OTHER,
				},
			},
		}, nil)

	api.ListTenantUsersHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []handlers.UserSearchResult
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, "tenant-1", resp[0].CompanyID)
	assert.True(t, resp[0].IsActive)
}

func TestInviteTenantUserHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	payload := map[string]string{
		"name":  "Tenant User",
		"email": "tenant@test.com",
		"role":  "viewer",
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("company_id", "tenant-1")
	c.Request, _ = http.NewRequest("POST", "/users/invitations", bytes.NewBuffer(body))

	mockCompany.On("CreateCompany", mock.Anything, &v1.CreateCompanyRequest{
		Name:       "Tenant User",
		OwnerEmail: "tenant@test.com",
	}).Return(&v1.CreateCompanyResponse{Id: "u2"}, nil)

	api.InviteTenantUserHandler(c)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestUpdateTenantUserRoleHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	body, _ := json.Marshal(map[string]string{"role": "operateur"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("company_id", "tenant-1")
	c.Params = gin.Params{{Key: "id", Value: "u1"}}
	c.Request, _ = http.NewRequest("PATCH", "/users/u1/role", bytes.NewBuffer(body))

	mockCompany.On("CreateCompany", mock.Anything, &v1.CreateCompanyRequest{Name: "u1"}).
		Return(&v1.CreateCompanyResponse{Id: "u1"}, nil)

	api.UpdateTenantUserRoleHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeactivateTenantUserHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("company_id", "tenant-1")
	c.Params = gin.Params{{Key: "id", Value: "u1"}}
	c.Request, _ = http.NewRequest("DELETE", "/users/u1", nil)

	mockCompany.On("CreateCompany", mock.Anything, &v1.CreateCompanyRequest{Name: "u1"}).
		Return(&v1.CreateCompanyResponse{Id: "u1"}, nil)

	api.DeactivateTenantUserHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateTenantUserStatusHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	body, _ := json.Marshal(map[string]bool{"is_active": true})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("company_id", "tenant-1")
	c.Params = gin.Params{{Key: "id", Value: "u1"}}
	c.Request, _ = http.NewRequest("PATCH", "/users/u1/status", bytes.NewBuffer(body))

	mockCompany.On("CreateCompany", mock.Anything, &v1.CreateCompanyRequest{Name: "u1"}).
		Return(&v1.CreateCompanyResponse{Id: "u1"}, nil)

	api.UpdateTenantUserStatusHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"is_active":true`)
}

func TestUpdateTenantUserStatusHandler_MapsPermissionError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	body, _ := json.Marshal(map[string]any{
		"is_active":  false,
		"company_id": "tenant-1",
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "u1"}}
	c.Request, _ = http.NewRequest("PATCH", "/users/u1/status", bytes.NewBuffer(body))

	mockCompany.On("CreateCompany", mock.Anything, &v1.CreateCompanyRequest{Name: "u1"}).
		Return(nil, status.Error(codes.PermissionDenied, "Cannot change your own account status"))

	api.UpdateTenantUserStatusHandler(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestListAuditLogsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/admin/audit-logs?limit=10&offset=5&company_id=c1", nil)

	mockCompany.On("ListAuditLogs", mock.Anything, mock.Anything).
		Return(&v1.ListAuditLogsResponse{
			Logs: []*v1.AuditLogEntry{
				{Id: "l1", Action: "test-action"},
			},
		}, nil)

	api.ListAuditLogsHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTeamStreamHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCompany := new(MockCompanyServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			CompanyService: mockCompany,
		},
	}

	mockStream := new(MockCompanyUpdateStream)
	mockStream.On("Recv").Return(&v1.WatchCompanyUpdatesResponse{
		EventType: "COMPANY_CREATED",
		EntityId:  "c1",
	}, nil).Once()
	mockStream.On("Recv").Return(nil, io.EOF).Once()

	mockCompany.On("WatchCompanyUpdates", mock.Anything, mock.Anything).
		Return(mockStream, nil)

	w := testutils.NewCloseNotifierRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/admin/teams/stream", nil)

	api.TeamStreamHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
