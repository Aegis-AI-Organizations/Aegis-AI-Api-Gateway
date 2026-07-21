package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/models"
	"github.com/gin-gonic/gin"
)

// GetVulnerabilitiesHandler handles GET /scans/:id/vulnerabilities
func (a *API) GetVulnerabilitiesHandler(c *gin.Context) {
	scanID := c.Param("id")
	if scanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scan id parameter is required"})
		return
	}

	grpcVulns, err := a.GRPCClient.GetVulnerabilities(c.Request.Context(), scanID)
	if err != nil {
		log.Printf("GRPC GetVulnerabilities error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	var vulns []models.Vulnerability
	for _, v := range grpcVulns {
		var exfiltratedData json.RawMessage
		if v.ExfiltratedData != "" {
			if json.Valid([]byte(v.ExfiltratedData)) {
				exfiltratedData = json.RawMessage(v.ExfiltratedData)
			} else {
				marshaledData, _ := json.Marshal(v.ExfiltratedData)
				exfiltratedData = json.RawMessage(marshaledData)
			}
		}

		var discoTime *time.Time
		if v.DiscoveredAt != nil {
			t := v.DiscoveredAt.AsTime()
			discoTime = &t
		}

		vulns = append(vulns, models.Vulnerability{
			ID:              v.Id,
			VulnType:        v.VulnType,
			Severity:        v.Severity,
			TargetEndpoint:  v.TargetEndpoint,
			Description:     v.Description,
			LootProof:       v.LootProof,
			ExfiltratedData: exfiltratedData,
			DiscoveredAt:    discoTime,
		})
	}

	if vulns == nil {
		vulns = []models.Vulnerability{}
	}

	c.JSON(http.StatusOK, vulns)
}
