package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/models"
)

// GetEvidencesHandler handles GET /vulnerabilities/{id}/evidences
func (a *API) GetEvidencesHandler(w http.ResponseWriter, r *http.Request) {
	vulnID := r.PathValue("id")
	if vulnID == "" {
		http.Error(w, "vulnerability id parameter is required", http.StatusBadRequest)
		return
	}

	grpcEvidences, err := a.GRPCClient.GetEvidences(r.Context(), vulnID)
	if err != nil {
		log.Printf("GRPC GetEvidences error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(evidences); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
