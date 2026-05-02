package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &handlers.API{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	api.HealthHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}

func TestRootHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &handlers.API{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	api.RootHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"service":"aegis-api-gateway","version":"pre-alpha"}`, w.Body.String())
}
