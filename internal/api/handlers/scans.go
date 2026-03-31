package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/models"
	"github.com/gin-gonic/gin"
)

func (a *API) GetScansHandler(c *gin.Context) {
	grpcScans, err := a.GRPCClient.ListScans(c.Request.Context())
	if err != nil {
		log.Printf("GRPC query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	var scans []models.Scan
	for _, s := range grpcScans {
		startStr := ""
		if s.StartedAt != nil {
			startStr = s.StartedAt.AsTime().Format(time.RFC3339)
		}

		var compStr *string
		if s.CompletedAt != nil {
			t := s.CompletedAt.AsTime().Format(time.RFC3339)
			compStr = &t
		}

		scans = append(scans, models.Scan{
			ID:                 s.ScanId,
			TemporalWorkflowID: s.TemporalWorkflowId,
			TargetImage:        s.TargetImage,
			Status:             s.Status,
			StartedAt:          startStr,
			CompletedAt:        compStr,
		})
	}

	if scans == nil {
		scans = []models.Scan{}
	}

	c.JSON(http.StatusOK, scans)
}

func (a *API) GetScanByIDHandler(c *gin.Context) {
	scanID := c.Param("id")
	if scanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scan id parameter is required"})
		return
	}

	s, err := a.GRPCClient.GetScanStatus(c.Request.Context(), scanID)
	if err != nil {
		log.Printf("GRPC GetScanStatus error: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Scan not found"})
		return
	}

	startStr := ""
	if s.StartedAt != nil {
		startStr = s.StartedAt.AsTime().Format(time.RFC3339)
	}
	var compStr *string
	if s.CompletedAt != nil {
		t := s.CompletedAt.AsTime().Format(time.RFC3339)
		compStr = &t
	}
	found := &models.Scan{
		ID:                 s.ScanId,
		TemporalWorkflowID: s.TemporalWorkflowId,
		TargetImage:        s.TargetImage,
		Status:             s.Status,
		StartedAt:          startStr,
		CompletedAt:        compStr,
	}

	c.JSON(http.StatusOK, found)
}
