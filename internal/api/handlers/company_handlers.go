package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

// OnboardCompanyHandler performs a full onboarding (Company + Owner + Token).
func (a *API) OnboardCompanyHandler(c *gin.Context) {
	var req struct {
		CompanyName   string `json:"company_name" binding:"required"`
		OwnerName     string `json:"owner_name" binding:"required"`
		OwnerEmail    string `json:"owner_email" binding:"required,email"`
		OwnerPassword string `json:"owner_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := a.GRPCClient.OnboardCompany(ctx, req.CompanyName, req.OwnerName, req.OwnerEmail, req.OwnerPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}
