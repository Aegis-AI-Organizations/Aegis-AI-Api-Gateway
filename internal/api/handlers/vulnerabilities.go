package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/models"
)

// GetVulnerabilitiesHandler handles GET /scans/{id}/vulnerabilities
func (a *API) GetVulnerabilitiesHandler(w http.ResponseWriter, r *http.Request) {
	scanID := r.PathValue("id")
	if scanID == "" {
		http.Error(w, "scan id parameter is required", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, vuln_type, severity, target_endpoint, description, discovered_at
		FROM vulnerabilities
		WHERE scan_id = $1
		ORDER BY discovered_at DESC
	`
	rows, err := a.DB.Query(query, scanID)
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

	var vulns []models.Vulnerability
	for rows.Next() {
		var v models.Vulnerability
		var endpoint, desc sql.NullString

		if err := rows.Scan(&v.ID, &v.VulnType, &v.Severity, &endpoint, &desc, &v.DiscoveredAt); err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}

		if endpoint.Valid {
			v.TargetEndpoint = endpoint.String
		}
		if desc.Valid {
			v.Description = desc.String
		}

		vulns = append(vulns, v)
	}

	if vulns == nil {
		vulns = []models.Vulnerability{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(vulns); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
