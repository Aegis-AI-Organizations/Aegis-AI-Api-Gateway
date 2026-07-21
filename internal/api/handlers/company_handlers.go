package handlers

import (
	"context"
	"net/http"
	"time"

	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc/aegis/v2"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type CurrentCompanyResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	OwnerID        string `json:"owner_id"`
	OwnerEmail     string `json:"owner_email"`
	MemberCount    int32  `json:"member_count"`
	AvatarURL      string `json:"avatar_url"`
	OrgSize        string `json:"org_size"`
	OrgType        string `json:"org_type"`
	TokenBalance   int64  `json:"token_balance"`
	DeploymentMode string `json:"deployment_mode"`
}

var organizationSizeByName = map[string]v1.OrganizationSize{
	"":                              v1.OrganizationSize_ORGANIZATION_SIZE_UNSPECIFIED,
	"ORGANIZATION_SIZE_UNSPECIFIED": v1.OrganizationSize_ORGANIZATION_SIZE_UNSPECIFIED,
	"ORGANIZATION_SIZE_1":           v1.OrganizationSize_ORGANIZATION_SIZE_1,
	"ORGANIZATION_SIZE_2_10":        v1.OrganizationSize_ORGANIZATION_SIZE_2_10,
	"ORGANIZATION_SIZE_11_50":       v1.OrganizationSize_ORGANIZATION_SIZE_11_50,
	"ORGANIZATION_SIZE_51_200":      v1.OrganizationSize_ORGANIZATION_SIZE_51_200,
	"ORGANIZATION_SIZE_201_500":     v1.OrganizationSize_ORGANIZATION_SIZE_201_500,
	"ORGANIZATION_SIZE_501_1000":    v1.OrganizationSize_ORGANIZATION_SIZE_501_1000,
	"ORGANIZATION_SIZE_1001_5000":   v1.OrganizationSize_ORGANIZATION_SIZE_1001_5000,
	"ORGANIZATION_SIZE_5001_10000":  v1.OrganizationSize_ORGANIZATION_SIZE_5001_10000,
	"ORGANIZATION_SIZE_10001_PLUS":  v1.OrganizationSize_ORGANIZATION_SIZE_10001_PLUS,
}

var organizationTypeByName = map[string]v1.OrganizationType{
	"":                              v1.OrganizationType_ORGANIZATION_TYPE_UNSPECIFIED,
	"ORGANIZATION_TYPE_UNSPECIFIED": v1.OrganizationType_ORGANIZATION_TYPE_UNSPECIFIED,
	"ORGANIZATION_TYPE_IT_SERVICES_AND_CONSULTING": v1.OrganizationType_ORGANIZATION_TYPE_IT_SERVICES_AND_CONSULTING,
	"ORGANIZATION_TYPE_SOFTWARE_DEVELOPMENT":       v1.OrganizationType_ORGANIZATION_TYPE_SOFTWARE_DEVELOPMENT,
	"ORGANIZATION_TYPE_FINANCIAL_SERVICES":         v1.OrganizationType_ORGANIZATION_TYPE_FINANCIAL_SERVICES,
	"ORGANIZATION_TYPE_HOSPITALS_AND_HEALTH_CARE":  v1.OrganizationType_ORGANIZATION_TYPE_HOSPITALS_AND_HEALTH_CARE,
	"ORGANIZATION_TYPE_RETAIL":                     v1.OrganizationType_ORGANIZATION_TYPE_RETAIL,
	"ORGANIZATION_TYPE_GOVERNMENT_ADMINISTRATION":  v1.OrganizationType_ORGANIZATION_TYPE_GOVERNMENT_ADMINISTRATION,
	"ORGANIZATION_TYPE_MANUFACTURING":              v1.OrganizationType_ORGANIZATION_TYPE_MANUFACTURING,
	"ORGANIZATION_TYPE_OTHER":                      v1.OrganizationType_ORGANIZATION_TYPE_OTHER,
}

func currentCompanyResponse(company *v1.CompanySummary) CurrentCompanyResponse {
	deploymentMode := "not_configured"
	if company.GetDeploymentToken() != "" {
		deploymentMode = "configured"
	}

	return CurrentCompanyResponse{
		ID:             company.GetId(),
		Name:           company.GetName(),
		OwnerID:        company.GetOwnerId(),
		OwnerEmail:     company.GetOwnerEmail(),
		MemberCount:    company.GetMemberCount(),
		AvatarURL:      company.GetAvatarUrl(),
		OrgSize:        company.GetOrgSize().String(),
		OrgType:        company.GetOrgType().String(),
		TokenBalance:   company.GetTokenBalance(),
		DeploymentMode: deploymentMode,
	}
}

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

// GetCurrentCompanyHandler returns the authenticated tenant company.
func (a *API) GetCurrentCompanyHandler(c *gin.Context) {
	if tenantCompanyID(c) == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Missing company scope"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	companies, err := a.GRPCClient.ListCompanies(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve current company"})
		return
	}
	if len(companies) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Current company not found"})
		return
	}

	c.JSON(http.StatusOK, currentCompanyResponse(companies[0]))
}

// UpdateCurrentCompanyHandler updates editable fields on the authenticated tenant company.
func (a *API) UpdateCurrentCompanyHandler(c *gin.Context) {
	if tenantCompanyID(c) == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Missing company scope"})
		return
	}

	var req struct {
		Name    string `json:"name" binding:"required"`
		OrgSize string `json:"org_size"`
		OrgType string `json:"org_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	orgSize, ok := organizationSizeByName[req.OrgSize]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid org_size"})
		return
	}
	orgType, ok := organizationTypeByName[req.OrgType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid org_type"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if _, err := a.GRPCClient.UpdateCurrentCompany(ctx, req.Name, orgSize, orgType); err != nil {
		writeCurrentCompanyGRPCError(c, err)
		return
	}

	companies, err := a.GRPCClient.ListCompanies(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Company updated but refresh failed"})
		return
	}
	if len(companies) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Current company not found"})
		return
	}

	c.JSON(http.StatusOK, currentCompanyResponse(companies[0]))
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

// AdminRevokeAgentTokenHandler invalidates the deployment token for a target company.
func (a *API) AdminRevokeAgentTokenHandler(c *gin.Context) {
	companyID := c.Param("id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company id is required"})
		return
	}

	a.revokeAgentToken(c, companyID)
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
	a.revokeAgentToken(c, "")
}

func (a *API) revokeAgentToken(c *gin.Context, companyID string) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if _, err := a.GRPCClient.RevokeAgentToken(ctx, companyID); err != nil {
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

func writeCurrentCompanyGRPCError(c *gin.Context, err error) {
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

	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update current company"})
}
