package handlers

import (
	"net/http"
	"strconv"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/types"
	"github.com/gin-gonic/gin"
)

// GetBalanceHandler handles GET /billing/balance or /admin/companies/:id/billing/balance
func (a *API) GetBalanceHandler(c *gin.Context) {
	companyID := c.Param("id")
	if companyID == "" {
		val, _ := c.Get(string(types.CompanyIDKey))
		companyID = val.(string)
	}

	resp, err := a.GRPCClient.GetBalance(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetLedgerHandler handles GET /billing/ledger or /admin/companies/:id/billing/ledger
func (a *API) GetLedgerHandler(c *gin.Context) {
	companyID := c.Param("id")
	if companyID == "" {
		val, _ := c.Get(string(types.CompanyIDKey))
		companyID = val.(string)
	}

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.ParseInt(limitStr, 10, 32)
	offset, _ := strconv.ParseInt(offsetStr, 10, 32)

	resp, err := a.GRPCClient.GetLedger(c.Request.Context(), companyID, int32(limit), int32(offset))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AdjustTokensHandler handles POST /admin/companies/:id/tokens/adjust
func (a *API) AdjustTokensHandler(c *gin.Context) {
	companyID := c.Param("id")

	var req struct {
		Amount int64  `json:"amount" binding:"required"`
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := a.GRPCClient.AdjustTokens(c.Request.Context(), companyID, req.Amount, req.Reason)
	if err != nil {
		// Brain returns FAILED_PRECONDITION for insufficient funds
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetUsageStatsHandler handles GET /billing/stats or /admin/companies/:id/billing/stats
func (a *API) GetUsageStatsHandler(c *gin.Context) {
	companyID := c.Param("id")
	if companyID == "" {
		val, _ := c.Get(string(types.CompanyIDKey))
		companyID = val.(string)
	}

	daysStr := c.DefaultQuery("days", "30")
	days, _ := strconv.ParseInt(daysStr, 10, 32)

	resp, err := a.GRPCClient.GetUsageStats(c.Request.Context(), companyID, int32(days))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
