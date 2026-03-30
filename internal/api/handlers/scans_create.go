package handlers

import (
	"log"
	"net/http"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/models"
	"github.com/gin-gonic/gin"
)

func (a *API) CreateScanHandler(c *gin.Context) {
	var req models.CreateScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	if req.TargetImage == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_image is required"})
		return
	}

	resp, err := a.GRPCClient.StartScan(c.Request.Context(), req.TargetImage)
	if err != nil {
		log.Printf("Failed to start scan via gRPC: %v", err)
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
