package handlers

import (
	"net/http"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/types"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func tenantCompanyID(c *gin.Context) string {
	if companyID, ok := c.Get(string(types.CompanyIDKey)); ok {
		if value, ok := companyID.(string); ok {
			return value
		}
	}
	return ""
}

func tenantUserErrorStatus(err error) int {
	switch status.Code(err) {
	case codes.NotFound:
		return http.StatusNotFound
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.InvalidArgument:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// ListTenantUsersHandler lists collaborators for the authenticated tenant.
func (a *API) ListTenantUsersHandler(c *gin.Context) {
	query := c.Query("search")
	companyID := tenantCompanyID(c)

	users, err := a.GRPCClient.SearchUsers(c.Request.Context(), query, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list collaborators"})
		return
	}

	results := make([]UserSearchResult, 0, len(users))
	for _, u := range users {
		results = append(results, UserSearchResult{
			ID:        u.Id,
			Name:      u.Name,
			Email:     u.OwnerEmail,
			Role:      u.DeploymentToken,
			CompanyID: u.OwnerId,
			AvatarURL: u.AvatarUrl,
			IsActive:  userSummaryIsActive(u),
		})
	}

	c.JSON(http.StatusOK, results)
}

// InviteTenantUserHandler invites a collaborator into the authenticated tenant.
func (a *API) InviteTenantUserHandler(c *gin.Context) {
	var req struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
		Role  string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := a.GRPCClient.InviteTenantUser(c.Request.Context(), req.Name, req.Email, req.Role, tenantCompanyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      resp.Id,
		"message": "Invitation collaborateur envoyée",
	})
}

// UpdateTenantUserRoleHandler updates a collaborator role inside the authenticated tenant.
func (a *API) UpdateTenantUserRoleHandler(c *gin.Context) {
	var req struct {
		Role      string `json:"role" binding:"required"`
		CompanyID string `json:"company_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID := c.Param("id")
	companyID := req.CompanyID
	if companyID == "" {
		companyID = tenantCompanyID(c)
	}
	resp, err := a.GRPCClient.UpdateTenantUserRole(c.Request.Context(), userID, req.Role, companyID)
	if err != nil {
		c.JSON(tenantUserErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      resp.Id,
		"role":    req.Role,
		"message": "Role updated",
	})
}

// UpdateTenantUserStatusHandler enables or disables a collaborator inside the authenticated tenant.
func (a *API) UpdateTenantUserStatusHandler(c *gin.Context) {
	var req struct {
		IsActive  bool   `json:"is_active"`
		CompanyID string `json:"company_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID := c.Param("id")
	companyID := req.CompanyID
	if companyID == "" {
		companyID = tenantCompanyID(c)
	}
	resp, err := a.GRPCClient.SetTenantUserActive(c.Request.Context(), userID, companyID, req.IsActive)
	if err != nil {
		c.JSON(tenantUserErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	message := "Collaborator deactivated"
	if req.IsActive {
		message = "Collaborator reactivated"
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        resp.Id,
		"is_active": req.IsActive,
		"message":   message,
	})
}

// DeactivateTenantUserHandler disables a collaborator inside the authenticated tenant.
func (a *API) DeactivateTenantUserHandler(c *gin.Context) {
	userID := c.Param("id")
	resp, err := a.GRPCClient.DeactivateTenantUser(c.Request.Context(), userID, tenantCompanyID(c))
	if err != nil {
		c.JSON(tenantUserErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      resp.Id,
		"message": "Collaborator deactivated",
	})
}
