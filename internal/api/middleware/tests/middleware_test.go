package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc/aegis/v2"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/middleware"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
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

type mockRedisWrapper struct {
	client *redis.Client
}

func (m *mockRedisWrapper) GetClient() *redis.Client {
	return m.client
}

func TestRedisRateLimiter_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock := redismock.NewClientMock()

	mock.ExpectIncr("ratelimit:127.0.0.1").SetVal(1)
	mock.ExpectExpire("ratelimit:127.0.0.1", 1*time.Minute).SetVal(true)

	r := gin.New()
	r.Use(middleware.RedisRateLimiter(&mockRedisWrapper{client: db}))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisRateLimiter_TooManyRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock := redismock.NewClientMock()

	mock.ExpectIncr("ratelimit:127.0.0.1").SetVal(101)

	r := gin.New()
	r.Use(middleware.RedisRateLimiter(&mockRedisWrapper{client: db}))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAgentRateLimiter_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock := redismock.NewClientMock()

	mock.ExpectIncr("ratelimit:agent:tenant1").SetVal(1)
	mock.ExpectExpire("ratelimit:agent:tenant1", 1*time.Second).SetVal(true)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(types.AgentTenantIDKey), "tenant1")
		c.Next()
	})
	r.Use(middleware.AgentRateLimiter(&mockRedisWrapper{client: db}))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAgentRateLimiter_TooManyRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock := redismock.NewClientMock()

	mock.ExpectIncr("ratelimit:agent:tenant1").SetVal(51)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(types.AgentTenantIDKey), "tenant1")
		c.Next()
	})
	r.Use(middleware.AgentRateLimiter(&mockRedisWrapper{client: db}))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisRateLimiter_RedisError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock := redismock.NewClientMock()

	mock.ExpectIncr("ratelimit:127.0.0.1").SetErr(fmt.Errorf("redis error"))

	r := gin.New()
	r.Use(middleware.RedisRateLimiter(&mockRedisWrapper{client: db}))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	r.ServeHTTP(w, req)

	// Should fallback to Next() on Redis error
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAgentRateLimiter_RedisError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock := redismock.NewClientMock()

	mock.ExpectIncr("ratelimit:agent:tenant1").SetErr(fmt.Errorf("redis error"))

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(types.AgentTenantIDKey), "tenant1")
		c.Next()
	})
	r.Use(middleware.AgentRateLimiter(&mockRedisWrapper{client: db}))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// Should fallback to Next() on Redis error
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRateLimiter_Local(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Test the local memory rate limiter
	r := gin.New()
	r.Use(middleware.RateLimiter(1, 1)) // 1 req/sec
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	// First request succeeds
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "1.1.1.1:1234"
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request fails (burst 1 exceeded)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "1.1.1.1:1234"
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}
