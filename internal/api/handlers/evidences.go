package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/models"
)

// GetEvidencesHandler handles GET /vulnerabilities/{id}/evidences
func (a *API) GetEvidencesHandler(w http.ResponseWriter, r *http.Request) {
	vulnID := r.PathValue("id")
	if vulnID == "" {
		http.Error(w, "vulnerability id parameter is required", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, vulnerability_id, payload_used, loot_data, captured_at
		FROM evidences
		WHERE vulnerability_id = $1
		ORDER BY captured_at DESC
	`
	rows, err := a.DB.Query(query, vulnID)
	if err != nil {
		log.Printf("DB query error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}()

	var evidences []models.Evidence
	for rows.Next() {
		var e models.Evidence
		var lootDataStr sql.NullString

		if err := rows.Scan(&e.ID, &e.VulnerabilityID, &e.PayloadUsed, &lootDataStr, &e.CapturedAt); err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}

		if lootDataStr.Valid {
			e.LootData = json.RawMessage(lootDataStr.String)
		}

		evidences = append(evidences, e)
	}

	if evidences == nil {
		evidences = []models.Evidence{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(evidences); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
