package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/models"
	"github.com/gin-gonic/gin"
)

// GetEvidencesHandler handles GET /vulnerabilities/:id/evidences
func (a *API) GetEvidencesHandler(c *gin.Context) {
	vulnID := c.Param("id")
	if vulnID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vulnerability id parameter is required"})
		return
	}

	grpcEvidences, err := a.GRPCClient.GetEvidences(c.Request.Context(), vulnID)
	if err != nil {
		log.Printf("GRPC GetEvidences error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	var evidences []models.Evidence
	for _, e := range grpcEvidences {
		var lootData json.RawMessage
		if e.LootData != "" {
			if json.Valid([]byte(e.LootData)) {
				lootData = json.RawMessage(e.LootData)
			} else {
				marshaledLoot, _ := json.Marshal(e.LootData)
				lootData = json.RawMessage(marshaledLoot)
			}
		}

		var capturedAt *time.Time
		if e.CapturedAt != nil {
			t := e.CapturedAt.AsTime()
			capturedAt = &t
		}

		evidences = append(evidences, models.Evidence{
			ID:              e.Id,
			VulnerabilityID: e.VulnerabilityId,
			PayloadUsed:     e.PayloadUsed,
			LootData:        lootData,
			CapturedAt:      capturedAt,
		})
	}

	if evidences == nil {
		evidences = []models.Evidence{}
	}

	c.JSON(http.StatusOK, evidences)
}
