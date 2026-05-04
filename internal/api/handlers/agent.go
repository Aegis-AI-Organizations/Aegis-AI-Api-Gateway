package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterAgentHandler handles the onboarding of a new agent using a deployment token.
func (a *API) RegisterAgentHandler(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
		Name  string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := a.GRPCClient.RegisterAgent(c.Request.Context(), req.Token, req.Name)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to register agent. Ensure your deployment token is valid."})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateAgentStatusHandler allows an agent to update its current operational state.
func (a *API) UpdateAgentStatusHandler(c *gin.Context) {
	agentID := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := a.GRPCClient.UpdateAgentStatus(c.Request.Context(), agentID, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update agent status"})
		return
	}

	// Update Redis Shadow Key for "Pro" health monitoring
	if a.Redis != nil {
		key := "agent:health:" + agentID
		// TTL of 90s (3x heartbeat interval of 30s)
		err := a.Redis.Client.Set(c.Request.Context(), key, "running", 90*1000000000).Err()
		if err != nil {
			// Log but don't fail the request
			log.Printf("⚠️ Failed to set agent health key in Redis: %v", err)
		}
	}

	c.JSON(http.StatusOK, resp)
}

// GetUploadLinkHandler generates a presigned MinIO URL for the agent to upload files.
func (a *API) GetUploadLinkHandler(c *gin.Context) {
	agentID := c.Param("id")
	filename := c.Query("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename query parameter is required"})
		return
	}

	resp, err := a.GRPCClient.GetUploadLink(c.Request.Context(), agentID, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate upload link"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
