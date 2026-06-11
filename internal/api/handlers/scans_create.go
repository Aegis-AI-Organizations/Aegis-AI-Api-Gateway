package handlers

import (
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/models"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/types"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	ipCount, apiCount, webappCount := scanBillingCounts(req)

	// 1. Billing Pre-flight Check
	check, err := a.GRPCClient.PreFlightCheck(c.Request.Context(), idStr, ipCount, apiCount, webappCount)
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
		log.Printf("Failed to process token consumption for company %s and target %s: %v", idStr, targetRef, err)
		c.JSON(tokenConsumptionErrorStatus(err), gin.H{"error": tokenConsumptionErrorMessage(err)})
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

func tokenConsumptionErrorStatus(err error) int {
	code := status.Code(err)
	switch code {
	case codes.FailedPrecondition:
		return http.StatusPaymentRequired
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unavailable, codes.DeadlineExceeded:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func tokenConsumptionErrorMessage(err error) string {
	st, ok := status.FromError(err)
	if !ok || st.Message() == "" {
		return "Failed to process token consumption"
	}

	switch st.Code() {
	case codes.FailedPrecondition:
		return st.Message()
	case codes.PermissionDenied:
		return "Token consumption is not allowed for this user"
	default:
		return "Failed to process token consumption"
	}
}

func scanBillingCounts(req models.CreateScanRequest) (int32, int32, int32) {
	ipCount := req.IpCount
	apiCount := req.ApiCount
	webappCount := req.WebappCount

	if strings.EqualFold(strings.TrimSpace(req.Scope), "topology") &&
		ipCount == 0 && apiCount == 0 && webappCount == 0 {
		targetIDs := scanTargetIDs(req)
		if len(targetIDs) > 0 {
			webappCount = int32(len(targetIDs))
		} else {
			webappCount = 1
		}
	}

	return ipCount, apiCount, webappCount
}

func scanTargetIDs(req models.CreateScanRequest) []string {
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
	return targetIDs
}

func scanTargetRef(req models.CreateScanRequest) string {
	if strings.TrimSpace(req.TargetImage) != "" {
		return strings.TrimSpace(req.TargetImage)
	}

	if strings.EqualFold(strings.TrimSpace(req.Scope), "topology") {
		targetIDs := scanTargetIDs(req)
		if len(targetIDs) == 0 {
			return "topology:all"
		}
		return "topology:" + strings.Join(targetIDs, ",")
	}

	return ""
}
