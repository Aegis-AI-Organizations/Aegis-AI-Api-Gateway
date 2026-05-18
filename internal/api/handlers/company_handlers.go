package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ListCompaniesHandler returns all registered companies (SuperAdmin only).
func (a *API) ListCompaniesHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	companies, err := a.GRPCClient.ListCompanies(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve companies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"companies": companies})
}

// CreateCompanyHandler registers a new company.
func (a *API) CreateCompanyHandler(c *gin.Context) {
	var req struct {
		Name       string `json:"name" binding:"required"`
		OwnerEmail string `json:"owner_email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := a.GRPCClient.CreateCompany(ctx, req.Name, req.OwnerEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company (owner may not exist or name is taken)"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// OnboardCompanyHandler creates a tenant and pending owner, then triggers first-login activation.
func (a *API) OnboardCompanyHandler(c *gin.Context) {
	var req struct {
		CompanyName string `json:"company_name" binding:"required"`
		OwnerName   string `json:"owner_name" binding:"required"`
		OwnerEmail  string `json:"owner_email" binding:"required,email"`
		PlanID      string `json:"plan_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Add plan to metadata if we can't update proto easily
	md := metadata.Pairs("x-plan-id", req.PlanID)
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := a.GRPCClient.OnboardCompany(ctx, req.CompanyName, req.OwnerName, req.OwnerEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// RotateAgentTokenHandler generates a fresh agent deployment token for the authenticated company.
func (a *API) RotateAgentTokenHandler(c *gin.Context) {
	a.rotateAgentToken(c, "")
}

// AdminRotateAgentTokenHandler generates a fresh agent deployment token for a target company.
func (a *API) AdminRotateAgentTokenHandler(c *gin.Context) {
	companyID := c.Param("id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company id is required"})
		return
	}

	a.rotateAgentToken(c, companyID)
}

func (a *API) rotateAgentToken(c *gin.Context, companyID string) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := a.GRPCClient.RotateAgentToken(ctx, companyID)
	if err != nil {
		writeAgentTokenGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"agent_token": resp.AgentToken})
}

// RevokeAgentTokenHandler invalidates the current agent deployment token.
func (a *API) RevokeAgentTokenHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if _, err := a.GRPCClient.RevokeAgentToken(ctx, ""); err != nil {
		writeAgentTokenGRPCError(c, err)
		return
	}

	c.AbortWithStatus(http.StatusNoContent)
}

func writeAgentTokenGRPCError(c *gin.Context, err error) {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.InvalidArgument:
			c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
			return
		case codes.PermissionDenied:
			c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
			return
		case codes.NotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
			return
		}
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update agent token"})
}
