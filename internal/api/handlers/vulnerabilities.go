package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/models"
)

// GetVulnerabilitiesHandler handles GET /scans/{id}/vulnerabilities
func (a *API) GetVulnerabilitiesHandler(w http.ResponseWriter, r *http.Request) {
	scanID := r.PathValue("id")
	if scanID == "" {
		http.Error(w, "scan id parameter is required", http.StatusBadRequest)
		return
	}

	grpcVulns, err := a.GRPCClient.GetVulnerabilities(r.Context(), scanID)
	if err != nil {
		log.Printf("GRPC GetVulnerabilities error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var vulns []models.Vulnerability
	for _, v := range grpcVulns {
		var discoTime *time.Time
		if v.DiscoveredAt != nil {
			t := v.DiscoveredAt.AsTime()
			discoTime = &t
		}

		vulns = append(vulns, models.Vulnerability{
			ID:             v.Id,
			VulnType:       v.VulnType,
			Severity:       v.Severity,
			TargetEndpoint: v.TargetEndpoint,
			Description:    v.Description,
			DiscoveredAt:   discoTime,
		})
	}

	if vulns == nil {
		vulns = []models.Vulnerability{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(vulns); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
