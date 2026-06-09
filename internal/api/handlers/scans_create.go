package handlers

import (
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/models"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/types"
	"github.com/gin-gonic/gin"
)

func (a *API) CreateScanHandler(c *gin.Context) {
	var req models.CreateScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	targetRef := scanTargetRef(req)
	if targetRef == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scan target is required"})
		return
	}

	companyID, _ := c.Get(string(types.CompanyIDKey))
	idStr := companyID.(string)

	// 1. Billing Pre-flight Check
	check, err := a.GRPCClient.PreFlightCheck(c.Request.Context(), idStr, req.IpCount, req.ApiCount, req.WebappCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Billing system unavailable"})
		return
	}

	if !check.SufficientBalance {
		c.JSON(http.StatusPaymentRequired, gin.H{
			"error":               "Insufficient token balance",
			"estimated_cost":      check.EstimatedCost,
			"current_balance":     check.CurrentBalance,
			"required_additional": check.EstimatedCost - check.CurrentBalance,
		})
		return
	}

	// 2. Deduct tokens
	_, err = a.GRPCClient.AdjustTokens(c.Request.Context(), idStr, -check.EstimatedCost, "Scan consumption: "+targetRef)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process token consumption"})
		return
	}

	// 3. Launch Scan
	resp, err := a.GRPCClient.StartScan(c.Request.Context(), targetRef)
	if err != nil {
		log.Printf("Failed to start scan via gRPC: %v", err)

		// Compensating action: Refund tokens
		_, refundErr := a.GRPCClient.AdjustTokens(c.Request.Context(), idStr, check.EstimatedCost, "Refund: scan launch failed")
		if refundErr != nil {
			log.Printf("CRITICAL: Failed to refund tokens for company %s after scan launch failure: %v", idStr, refundErr)
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start workflow orchestrator"})
		return
	}

	log.Printf("Started Orchestration Workflow for scanID: %s", resp.ScanId)

	res := models.CreateScanResponse{
		ScanID: resp.ScanId,
		Status: resp.Status,
	}

	c.JSON(http.StatusCreated, res)
}

func scanTargetRef(req models.CreateScanRequest) string {
	if strings.TrimSpace(req.TargetImage) != "" {
		return strings.TrimSpace(req.TargetImage)
	}

	if strings.EqualFold(strings.TrimSpace(req.Scope), "topology") {
		targetIDs := make([]string, 0, len(req.TargetNodeIDs))
		seen := map[string]struct{}{}
		for _, targetID := range req.TargetNodeIDs {
			normalized := strings.TrimSpace(targetID)
			if normalized == "" {
				continue
			}
			if _, exists := seen[normalized]; exists {
				continue
			}
			seen[normalized] = struct{}{}
			targetIDs = append(targetIDs, normalized)
		}
		sort.Strings(targetIDs)
		if len(targetIDs) == 0 {
			return "topology:all"
		}
		return "topology:" + strings.Join(targetIDs, ",")
	}

	return ""
}
