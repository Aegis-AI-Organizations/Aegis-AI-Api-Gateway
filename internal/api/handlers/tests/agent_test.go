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
	"google.golang.org/grpc"
)

type MockAgentServiceClient struct {
	mock.Mock
}

func (m *MockAgentServiceClient) RegisterAgent(ctx context.Context, in *v1.RegisterAgentRequest, opts ...grpc.CallOption) (*v1.RegisterAgentResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.RegisterAgentResponse), args.Error(1)
}

func (m *MockAgentServiceClient) UpdateAgentStatus(ctx context.Context, in *v1.UpdateAgentStatusRequest, opts ...grpc.CallOption) (*v1.UpdateAgentStatusResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.UpdateAgentStatusResponse), args.Error(1)
}

func (m *MockAgentServiceClient) GetUploadLink(ctx context.Context, in *v1.GetUploadLinkRequest, opts ...grpc.CallOption) (*v1.GetUploadLinkResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.GetUploadLinkResponse), args.Error(1)
}

func (m *MockAgentServiceClient) VerifyAgentSecret(ctx context.Context, in *v1.VerifyAgentSecretRequest, opts ...grpc.CallOption) (*v1.VerifyAgentSecretResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.VerifyAgentSecretResponse), args.Error(1)
}

func TestRegisterAgentHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAgent := new(MockAgentServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			AgentService: mockAgent,
		},
	}

	payload := map[string]string{"token": "ag_token", "name": "Agent1"}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/agents/register", bytes.NewBuffer(body))

	mockAgent.On("RegisterAgent", mock.Anything, &v1.RegisterAgentRequest{Token: "ag_token", Name: "Agent1"}).
		Return(&v1.RegisterAgentResponse{AgentId: "a1", AgentSecret: "s1"}, nil)

	api.RegisterAgentHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegisterAgentHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAgent := new(MockAgentServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			AgentService: mockAgent,
		},
	}

	payload := map[string]string{"token": "ag_token"}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/agents/register", bytes.NewBuffer(body))

	mockAgent.On("RegisterAgent", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("grpc error"))

	api.RegisterAgentHandler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateAgentStatusHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAgent := new(MockAgentServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			AgentService: mockAgent,
		},
	}

	payload := map[string]string{"status": "IDLE"}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "a1"}}
	c.Request, _ = http.NewRequest("POST", "/agents/a1/status", bytes.NewBuffer(body))

	mockAgent.On("UpdateAgentStatus", mock.Anything, &v1.UpdateAgentStatusRequest{AgentId: "a1", Status: "IDLE"}).
		Return(&v1.UpdateAgentStatusResponse{Success: true}, nil)

	api.UpdateAgentStatusHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateAgentStatusHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAgent := new(MockAgentServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			AgentService: mockAgent,
		},
	}

	payload := map[string]string{"status": "IDLE"}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "a1"}}
	c.Request, _ = http.NewRequest("POST", "/agents/a1/status", bytes.NewBuffer(body))

	mockAgent.On("UpdateAgentStatus", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("grpc error"))

	api.UpdateAgentStatusHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetUploadLinkHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAgent := new(MockAgentServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			AgentService: mockAgent,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "a1"}}
	c.Request, _ = http.NewRequest("GET", "/agents/a1/upload-url?filename=test.txt", nil)

	mockAgent.On("GetUploadLink", mock.Anything, &v1.GetUploadLinkRequest{AgentId: "a1", Filename: "test.txt"}).
		Return(&v1.GetUploadLinkResponse{Url: "http://minio/upload", Method: "PUT"}, nil)

	api.GetUploadLinkHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUploadLinkHandler_NoFilename(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &handlers.API{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/agents/a1/upload-url", nil)

	api.GetUploadLinkHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUploadLinkHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAgent := new(MockAgentServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			AgentService: mockAgent,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "a1"}}
	c.Request, _ = http.NewRequest("GET", "/agents/a1/upload-url?filename=test.txt", nil)

	mockAgent.On("GetUploadLink", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("grpc error"))

	api.GetUploadLinkHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
