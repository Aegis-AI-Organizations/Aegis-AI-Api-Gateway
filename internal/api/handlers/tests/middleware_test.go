package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestAgentAuthMiddleware_Unit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := new(MockInternalAuthClient)
	client := &agrpc.Client{
		InternalAuthService: mockAuth,
	}

	t.Run("NoToken", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/test", nil)

		middleware.AgentAuthMiddleware(client, nil)(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.True(t, c.IsAborted())
	})

	t.Run("Success", func(t *testing.T) {
		mockAuth.On("VerifyToken", mock.Anything, &v1.VerifyTokenRequest{Token: "valid-token"}).
			Return(&v1.VerifyTokenResponse{Valid: true, TenantId: "t1"}, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer valid-token")

		middleware.AgentAuthMiddleware(client, nil)(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.False(t, c.IsAborted())
	})

	t.Run("BrainFailure", func(t *testing.T) {
		mockAuth.On("VerifyToken", mock.Anything, &v1.VerifyTokenRequest{Token: "invalid-token"}).
			Return(&v1.VerifyTokenResponse{Valid: false}, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer invalid-token")

		middleware.AgentAuthMiddleware(client, nil)(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.True(t, c.IsAborted())
	})

	t.Run("GRPCError", func(t *testing.T) {
		mockAuth.On("VerifyToken", mock.Anything, &v1.VerifyTokenRequest{Token: "error-token"}).
			Return(nil, fmt.Errorf("grpc error")).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer error-token")

		middleware.AgentAuthMiddleware(client, nil)(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.True(t, c.IsAborted())
	})
}

func TestRedisRateLimiter_Unit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("NilClient", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/test", nil)

		middleware.RedisRateLimiter(nil)(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRateLimiter_Unit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)

	middleware.RateLimiter(100, 100)(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_Unit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	t.Run("NoAuth", func(t *testing.T) {
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request, _ = http.NewRequest("GET", "/test", nil)

        middleware.AuthMiddleware()(c)

        assert.Equal(t, http.StatusUnauthorized, w.Code)
        assert.True(t, c.IsAborted())
    })
}
