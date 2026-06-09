package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInstallScriptHandler_WithToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &handlers.API{}
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.GET("/install.sh", api.InstallScriptHandler)
	req, _ := http.NewRequest("GET", "/install.sh?token=ag_test12345", nil)
	c.Request = req
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/x-shellscript", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "DEPLOYMENT_TOKEN=ag_test12345")
	assert.Contains(t, w.Body.String(), "curl -fsSL")
	assert.Contains(t, w.Body.String(), "releases/latest/download/aegis-ai-agent")
	assert.Contains(t, w.Body.String(), "storage.aegis-ai.fr/releases/aegis-ai-agent")
}

func TestInstallScriptHandler_WithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &handlers.API{}
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.GET("/install.sh", api.InstallScriptHandler)
	req, _ := http.NewRequest("GET", "/install.sh", nil)
	c.Request = req
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "missing or invalid agent deployment token")
}
