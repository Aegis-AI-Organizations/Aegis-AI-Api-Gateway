package handlers

import (
	"net/http"
	"strconv"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// GetBalanceHandler handles GET /billing/balance
func (a *API) GetBalanceHandler(c *gin.Context) {
	companyID, _ := c.Get(string(middleware.CompanyIDKey))
	idStr := companyID.(string)

	resp, err := a.GRPCClient.GetBalance(c.Request.Context(), idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetLedgerHandler handles GET /billing/ledger
func (a *API) GetLedgerHandler(c *gin.Context) {
	companyID, _ := c.Get(string(middleware.CompanyIDKey))
	idStr := companyID.(string)

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.ParseInt(limitStr, 10, 32)
	offset, _ := strconv.ParseInt(offsetStr, 10, 32)

	resp, err := a.GRPCClient.GetLedger(c.Request.Context(), idStr, int32(limit), int32(offset))
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
