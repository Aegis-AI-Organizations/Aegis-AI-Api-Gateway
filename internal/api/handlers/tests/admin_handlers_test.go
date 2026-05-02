package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc/aegis/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
