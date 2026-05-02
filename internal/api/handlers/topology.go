package handlers

import (
	"net/http"

	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc/aegis/v2"
	"github.com/gin-gonic/gin"
)

// ReportTopologyHandler receives discovered infrastructure data from Agents and sends it to the Brain.
func (a *API) ReportTopologyHandler(c *gin.Context) {
	var topology v1.NetworkTopology
	if err := c.ShouldBindJSON(&topology); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid topology format. Please check the schema for hosts, containers, and processes."})
		return
	}

	resp, err := a.GRPCClient.ReportTopology(c.Request.Context(), &topology)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to report topology to the backend service"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
