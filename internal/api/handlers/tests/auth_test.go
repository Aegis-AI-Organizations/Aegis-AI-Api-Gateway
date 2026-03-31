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
	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/grpc"
	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/grpc/aegis/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

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

func TestLoginHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := new(MockAuthServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			AuthService: mockAuth,
		},
	}

	payload := map[string]string{"email": "test@example.com", "password": "password"}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))

	mockAuth.On("Login", mock.Anything, &v1.LoginRequest{Email: "test@example.com", Password: "password"}).
		Return(&v1.LoginResponse{AccessToken: "access", RefreshToken: "refresh"}, nil)

	api.LoginHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "access", resp["access_token"])

	// Check cookie
	cookies := w.Result().Cookies()
	var found bool
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			assert.Equal(t, "refresh", cookie.Value)
			assert.True(t, cookie.HttpOnly)
			assert.False(t, cookie.Secure) // TestMode is not ReleaseMode
			found = true
		}
	}
	assert.True(t, found)
}

func TestLoginHandler_InvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &handlers.API{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString("invalid json"))

	api.LoginHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoginHandler_GRPCError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := new(MockAuthServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			AuthService: mockAuth,
		},
	}

	payload := map[string]string{"email": "test@example.com", "password": "wrong"}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))

	mockAuth.On("Login", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("invalid credentials"))

	api.LoginHandler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRefreshHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := new(MockAuthServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			AuthService: mockAuth,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/auth/refresh", nil)
	c.Request.AddCookie(&http.Cookie{Name: "refresh_token", Value: "old-refresh"})

	mockAuth.On("Refresh", mock.Anything, &v1.RefreshRequest{RefreshToken: "old-refresh"}).
		Return(&v1.RefreshResponse{AccessToken: "new-access"}, nil)

	api.RefreshHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "new-access", resp["access_token"])
}

func TestRefreshHandler_MissingCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &handlers.API{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/auth/refresh", nil)

	api.RefreshHandler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogoutHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := new(MockAuthServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			AuthService: mockAuth,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/auth/logout", nil)
	c.Request.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh-to-clear"})

	mockAuth.On("Logout", mock.Anything, &v1.LogoutRequest{RefreshToken: "refresh-to-clear"}).
		Return(&v1.LogoutResponse{Success: true}, nil)

	api.LogoutHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check cookie cleared
	cookies := w.Result().Cookies()
	var found bool
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			assert.Equal(t, "", cookie.Value)
			assert.Equal(t, -1, cookie.MaxAge)
			found = true
		}
	}
	assert.True(t, found)
}

func TestRefreshHandler_GRPCError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := new(MockAuthServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			AuthService: mockAuth,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/auth/refresh", nil)
	c.Request.AddCookie(&http.Cookie{Name: "refresh_token", Value: "invalid-token"})

	mockAuth.On("Refresh", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("grpc error"))

	api.RefreshHandler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogoutHandler_MissingCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &handlers.API{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/auth/logout", nil)

	api.LogoutHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
