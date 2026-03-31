package handlers

import (
	"net/http"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	"github.com/gin-gonic/gin"
)

// API holds the core dependencies dynamically injected by our server initialization.
type API struct {
	GRPCClient *agrpc.Client
}

// HealthHandler returns a simple 200 OK status for Kubernetes liveness probes.
func (a *API) HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// RootHandler returns the service name and version.
func (a *API) RootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "aegis-api-gateway",
		"version": "pre-alpha",
	})
}
