package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc/aegis/v2"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// MockInternalAuthClient
type MockInternalAuthClient struct {
	mock.Mock
}

func (m *MockInternalAuthClient) VerifyToken(ctx context.Context, in *v1.VerifyTokenRequest, opts ...grpc.CallOption) (*v1.VerifyTokenResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.VerifyTokenResponse), args.Error(1)
}

func TestAgentAuthMiddleware_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.AgentAuthMiddleware(nil, nil))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAgentAuthMiddleware_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := new(MockInternalAuthClient)
	client := &agrpc.Client{
		InternalAuthService: mockAuth,
	}

	mockAuth.On("VerifyToken", mock.Anything, &v1.VerifyTokenRequest{Token: "valid-token"}).
		Return(&v1.VerifyTokenResponse{Valid: true, TenantId: "t1"}, nil)

	r := gin.New()
	r.Use(middleware.AgentAuthMiddleware(client, nil))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRedisRateLimiter_NoRedis(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RedisRateLimiter(nil))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
