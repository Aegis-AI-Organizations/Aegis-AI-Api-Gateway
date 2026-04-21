package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/middleware"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_NoSecret(t *testing.T) {
	os.Setenv("JWT_SECRET", "")
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthMiddleware_Success(t *testing.T) {
	secret := "test-secret"
	os.Setenv("JWT_SECRET", secret)
	gin.SetMode(gin.TestMode)

	// Create a valid token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":        "u1",
		"company_id": "c1",
		"role":       "admin",
		"exp":        time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(secret))

	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		assert.Equal(t, "u1", userID)
		c.Status(http.StatusOK)
	})

	// Test header
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test query param
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test?token="+tokenString, nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_Errors(t *testing.T) {
	secret := "test-secret"
	os.Setenv("JWT_SECRET", secret)
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// No token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Invalid token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Missing claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "u1",
		// missing company_id and role
	})
	tokenString, _ := token.SignedString([]byte(secret))
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
