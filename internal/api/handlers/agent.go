package handlers

import (
	"log"
	"net/http"
	"time"

	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc/aegis/v2"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type agentRecordResponse struct {
	ID        string     `json:"id"`
	CompanyID string     `json:"company_id"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	LastSeen  *time.Time `json:"last_seen"`
	CreatedAt *time.Time `json:"created_at"`
}

type agentStatusSummaryResponse struct {
	TotalAgents    int32      `json:"total_agents"`
	ActiveAgents   int32      `json:"active_agents"`
	InactiveAgents int32      `json:"inactive_agents"`
	LastSeen       *time.Time `json:"last_seen"`
}

func timestampToTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func agentRecordToResponse(agent *v1.AgentRecord) agentRecordResponse {
	return agentRecordResponse{
		ID:        agent.GetId(),
		CompanyID: agent.GetCompanyId(),
		Name:      agent.GetName(),
		Status:    agent.GetStatus(),
		LastSeen:  timestampToTime(agent.GetLastSeen()),
		CreatedAt: timestampToTime(agent.GetCreatedAt()),
	}
}

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

// ListAgentsHandler returns agents visible from the authenticated company scope.
func (a *API) ListAgentsHandler(c *gin.Context) {
	companyID := c.Query("company_id")

	resp, err := a.GRPCClient.ListAgents(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list agents"})
		return
	}

	agents := make([]agentRecordResponse, 0, len(resp.GetAgents()))
	for _, agent := range resp.GetAgents() {
		agents = append(agents, agentRecordToResponse(agent))
	}

	c.JSON(http.StatusOK, gin.H{"agents": agents})
}

// GetAgentStatusSummaryHandler returns aggregated agent health for the authenticated company scope.
func (a *API) GetAgentStatusSummaryHandler(c *gin.Context) {
	companyID := c.Query("company_id")

	resp, err := a.GRPCClient.GetAgentStatusSummary(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agent status summary"})
		return
	}

	c.JSON(http.StatusOK, agentStatusSummaryResponse{
		TotalAgents:    resp.GetTotalAgents(),
		ActiveAgents:   resp.GetActiveAgents(),
		InactiveAgents: resp.GetInactiveAgents(),
		LastSeen:       timestampToTime(resp.GetLastSeen()),
	})
}
