package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListAuditLogsHandler retrieves system audit logs via Brain gRPC.
func (a *API) ListAuditLogsHandler(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	companyID := c.Query("company_id")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	resp, err := a.GRPCClient.ListAuditLogs(c.Request.Context(), int32(limit), int32(offset), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit logs"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

type CompanySearchResult struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	DeploymentToken string `json:"deployment_token"`
	OwnerEmail      string `json:"owner_email"`
	AvatarURL       string `json:"avatar_url"`
}

type UserSearchResult struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CompanyID string `json:"company_id"`
	AvatarURL string `json:"avatar_url"`
}

// SearchCompaniesHandler searches for companies by name or ID via Brain gRPC.
func (a *API) SearchCompaniesHandler(c *gin.Context) {
	query := c.Query("search")

	companies, err := a.GRPCClient.SearchCompanies(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search companies via Brain"})
		return
	}

	results := make([]CompanySearchResult, 0, len(companies))
	for _, res := range companies {
		results = append(results, CompanySearchResult{
			ID:              res.Id,
			Name:            res.Name,
			DeploymentToken: res.DeploymentToken,
			OwnerEmail:      res.OwnerEmail,
			AvatarURL:       res.AvatarUrl,
		})
	}

	c.JSON(http.StatusOK, results)
}

// SearchUsersHandler searches for users by name, email, or ID via Brain gRPC.
func (a *API) SearchUsersHandler(c *gin.Context) {
	query := c.Query("search")
	companyID := c.Query("company_id")

	users, err := a.GRPCClient.SearchUsers(c.Request.Context(), query, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search users via Brain"})
		return
	}

	results := make([]UserSearchResult, 0, len(users))
	for _, u := range users {
		results = append(results, UserSearchResult{
			ID:        u.Id,
			Name:      u.Name,
			Email:     u.OwnerEmail,      // mapped from response
			Role:      u.DeploymentToken, // mapped from response
			CompanyID: u.OwnerId,         // mapped from response
			AvatarURL: u.AvatarUrl,
		})
	}

	c.JSON(http.StatusOK, results)
}

// CreateUserHandler invites a new user via Brain gRPC proxy.
func (a *API) CreateUserHandler(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Email     string `json:"email" binding:"required,email"`
		Role      string `json:"role" binding:"required"`
		CompanyID string `json:"company_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := a.GRPCClient.AdminCreateUser(c.Request.Context(), req.Name, req.Email, req.Role, req.CompanyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      resp.Id,
		"message": "Invitation collaborateur envoyée",
	})
}
