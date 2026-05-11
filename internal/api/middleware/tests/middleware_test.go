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
	db_internal "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/db"
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

// MockAgentClient
type MockAgentClient struct {
	mock.Mock
	v1.AgentServiceClient
}

func (m *MockAgentClient) VerifyAgentSecret(ctx context.Context, in *v1.VerifyAgentSecretRequest, opts ...grpc.CallOption) (*v1.VerifyAgentSecretResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.VerifyAgentSecretResponse), args.Error(1)
}

func (m *MockAgentClient) RegisterAgent(ctx context.Context, in *v1.RegisterAgentRequest, opts ...grpc.CallOption) (*v1.RegisterAgentResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*v1.RegisterAgentResponse), args.Error(1)
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

func TestAgentAuthMiddleware_Register_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := new(MockInternalAuthClient)
	client := &agrpc.Client{
		InternalAuthService: mockAuth,
	}

	mockAuth.On("VerifyToken", mock.Anything, &v1.VerifyTokenRequest{Token: "ag_0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefg"}).
		Return(&v1.VerifyTokenResponse{Valid: true, TenantId: "t1"}, nil)

	r := gin.New()
	r.Use(middleware.AgentAuthMiddleware(client, nil))
	r.POST("/api/agents/register", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/agents/register", nil)
	req.Header.Set("Authorization", "Bearer ag_0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefg")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAgentAuthMiddleware_Operation_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAgent := new(MockAgentClient)
	client := &agrpc.Client{
		AgentService: mockAgent,
	}

	mockAgent.On("VerifyAgentSecret", mock.Anything, &v1.VerifyAgentSecretRequest{AgentId: "a1", Secret: "s1"}).
		Return(&v1.VerifyAgentSecretResponse{Valid: true, TenantId: "t1"}, nil)

	r := gin.New()
	r.Use(middleware.AgentAuthMiddleware(client, nil))
	r.GET("/api/agents/:id/status", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/agents/a1/status", nil)
	req.Header.Set("Authorization", "Bearer s1")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAgentAuthMiddleware_DeploymentToken_On_Operation_Rejection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.AgentAuthMiddleware(nil, nil))
	r.GET("/api/agents/:id/status", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/agents/a1/status", nil)
	req.Header.Set("Authorization", "Bearer ag_0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefg") // Starts with ag_
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
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

func TestAgentAuthMiddleware_Redis_Hit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, rMock := redismock.NewClientMock()
	redisClient := &db_internal.RedisClient{Client: db}

	// Mock Redis hit
	rMock.ExpectGet("agent_token:ag_0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefg").SetVal("tenant1")

	r := gin.New()
	r.Use(middleware.AgentAuthMiddleware(nil, redisClient))
	r.POST("/api/agents/register", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/agents/register", nil)
	req.Header.Set("Authorization", "Bearer ag_0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefg")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, rMock.ExpectationsWereMet())
}

func TestAgentAuthMiddleware_Redis_Miss_Valid_Secret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, rMock := redismock.NewClientMock()
	redisClient := &db_internal.RedisClient{Client: db}
	mockAgent := new(MockAgentClient)
	client := &agrpc.Client{AgentService: mockAgent}

	// Mock Redis miss
	rMock.ExpectGet("agent_secret:a1:s1").RedisNil()
	// Mock gRPC success
	mockAgent.On("VerifyAgentSecret", mock.Anything, &v1.VerifyAgentSecretRequest{AgentId: "a1", Secret: "s1"}).
		Return(&v1.VerifyAgentSecretResponse{Valid: true, TenantId: "t1"}, nil)
	// Mock Redis set
	rMock.ExpectSet("agent_secret:a1:s1", "t1", 30*time.Minute).SetVal("OK")

	r := gin.New()
	r.Use(middleware.AgentAuthMiddleware(client, redisClient))
	r.GET("/api/agents/:id/status", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/agents/a1/status", nil)
	req.Header.Set("Authorization", "Bearer s1")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, rMock.ExpectationsWereMet())
}

func TestAgentAuthMiddleware_DeploymentToken_On_Operation_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.AgentAuthMiddleware(nil, nil))
	r.GET("/api/agents/:id/status", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/agents/a1/status", nil)
	req.Header.Set("Authorization", "Bearer ag_0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefg")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAgentAuthMiddleware_Secret_On_Register_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.AgentAuthMiddleware(nil, nil))
	r.POST("/api/agents/register", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/agents/register", nil)
	req.Header.Set("Authorization", "Bearer secret123") // No ag_ prefix
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAgentAuthMiddleware_MalformedDeploymentToken_On_Register_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.AgentAuthMiddleware(nil, nil))
	r.POST("/api/agents/register", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/agents/register", nil)
	req.Header.Set("Authorization", "Bearer ag_too-short")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAgentAuthMiddleware_GRPC_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := new(MockInternalAuthClient)
	client := &agrpc.Client{InternalAuthService: mockAuth}

	mockAuth.On("VerifyToken", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("grpc error"))

	r := gin.New()
	r.POST("/api/agents/register", middleware.AgentAuthMiddleware(client, nil), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/agents/register", nil)
	req.Header.Set("Authorization", "Bearer ag_0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefg")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAgentAuthMiddleware_No_ID_Param(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.AgentAuthMiddleware(nil, nil))
	r.GET("/api/status", func(c *gin.Context) { c.Status(200) }) // Route without :id

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/status", nil)
	req.Header.Set("Authorization", "Bearer secret123")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
