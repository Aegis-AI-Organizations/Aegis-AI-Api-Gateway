package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc/aegis/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdateProfileHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := new(MockAuthServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			AuthService: mockAuth,
		},
	}

	payload := map[string]string{"name": "New Name"}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("PUT", "/users/me/profile", bytes.NewBuffer(body))

	mockAuth.On("UpdateProfile", mock.Anything, mock.Anything).
		Return(&v1.UpdateProfileResponse{Success: true}, nil)

	api.UpdateProfileHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateProfileHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := new(MockAuthServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			AuthService: mockAuth,
		},
	}

	payload := map[string]string{"name": "New Name"}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("PUT", "/users/me/profile", bytes.NewBuffer(body))

	mockAuth.On("UpdateProfile", mock.Anything, mock.Anything).
		Return(nil, status.Error(codes.Internal, "internal error"))

	api.UpdateProfileHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateEmailHandler_Conflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := new(MockAuthServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			AuthService: mockAuth,
		},
	}

	payload := map[string]string{"email": "existing@example.com"}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("PUT", "/users/me/email", bytes.NewBuffer(body))

	mockAuth.On("UpdateEmail", mock.Anything, &v1.UpdateEmailRequest{NewEmail: "existing@example.com"}).
		Return(nil, status.Error(codes.AlreadyExists, "email in use"))

	api.UpdateEmailHandler(c)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestUpdateEmailHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := new(MockAuthServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			AuthService: mockAuth,
		},
	}

	payload := map[string]string{"email": "new@example.com"}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("PUT", "/users/me/email", bytes.NewBuffer(body))

	mockAuth.On("UpdateEmail", mock.Anything, &v1.UpdateEmailRequest{NewEmail: "new@example.com"}).
		Return(&v1.UpdateEmailResponse{Success: true}, nil)

	api.UpdateEmailHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdatePasswordHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := new(MockAuthServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			AuthService: mockAuth,
		},
	}

	payload := map[string]string{
		"old_password": "correct",
		"new_password": "verysecret123",
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("PUT", "/users/me/password", bytes.NewBuffer(body))

	mockAuth.On("UpdatePassword", mock.Anything, &v1.UpdatePasswordRequest{
		OldPassword: "correct",
		NewPassword: "verysecret123",
	}).Return(&v1.UpdatePasswordResponse{Success: true}, nil)

	api.UpdatePasswordHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdatePasswordHandler_WrongPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := new(MockAuthServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			AuthService: mockAuth,
		},
	}

	payload := map[string]string{
		"old_password": "wrong",
		"new_password": "verysecret123",
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("PUT", "/users/me/password", bytes.NewBuffer(body))

	mockAuth.On("UpdatePassword", mock.Anything, &v1.UpdatePasswordRequest{
		OldPassword: "wrong",
		NewPassword: "verysecret123",
	}).Return(nil, status.Error(codes.Unauthenticated, "wrong password"))

	api.UpdatePasswordHandler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdatePasswordHandler_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &handlers.API{}

	payload := map[string]string{"old_password": "any", "new_password": "short"}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("PUT", "/users/me/password", bytes.NewBuffer(body))

	api.UpdatePasswordHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
