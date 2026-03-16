package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/models"
)

// CreateScanHandler handles POST /scans
func (a *API) CreateScanHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method must be POST", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.TargetImage == "" {
		http.Error(w, "target_image is required", http.StatusBadRequest)
		return
	}

	scanID := uuid.New().String()
	workflowID := fmt.Sprintf("pentest-workflow-%s", scanID)

	query := `INSERT INTO scans (id, temporal_workflow_id, target_image, status) VALUES ($1, $2, $3, 'PENDING')`
	_, err := a.DB.Exec(query, scanID, workflowID, req.TargetImage)
	if err != nil {
		log.Printf("Failed to insert scan into DB: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "BRAIN_TASK_QUEUE",
	}

	we, err := a.TemporalClient.ExecuteWorkflow(context.Background(), options, "PentestWorkflow", scanID, req.TargetImage)
	if err != nil {
		log.Printf("Failed to start Temporal workflow: %v", err)
		if _, dbErr := a.DB.Exec(`UPDATE scans SET status = 'FAILED' WHERE id = $1`, scanID); dbErr != nil {
			log.Printf("Failed to update scan status to FAILED: %v", dbErr)
		}

		http.Error(w, "Failed to start workflow orchestrator", http.StatusInternalServerError)
		return
	}

	log.Printf("Started Orchestration Workflow. WorkflowID: %s, RunID: %s", we.GetID(), we.GetRunID())

	res := models.CreateScanResponse{
		ScanID:             scanID,
		TemporalWorkflowID: workflowID,
		Status:             "PENDING",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// GetScansHandler handles GET /scans
func (a *API) GetScansHandler(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT id, temporal_workflow_id, target_image, status, started_at, completed_at
		FROM scans
		ORDER BY started_at DESC
	`
	rows, err := a.DB.Query(query)
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

	var scans []models.Scan
	for rows.Next() {
		var s models.Scan
		var completedAt sql.NullString

		if err := rows.Scan(&s.ID, &s.TemporalWorkflowID, &s.TargetImage, &s.Status, &s.StartedAt, &completedAt); err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}

		if completedAt.Valid {
			s.CompletedAt = &completedAt.String
		}

		scans = append(scans, s)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if scans == nil {
		scans = []models.Scan{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(scans); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// GetScanByIDHandler handles GET /scans/{id}
func (a *API) GetScanByIDHandler(w http.ResponseWriter, r *http.Request) {
	scanID := r.PathValue("id")
	if scanID == "" {
		http.Error(w, "scan id parameter is required", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, temporal_workflow_id, target_image, status, started_at, completed_at
		FROM scans
		WHERE id = $1
	`

	var s models.Scan
	var completedAt sql.NullString

	err := a.DB.QueryRow(query, scanID).Scan(&s.ID, &s.TemporalWorkflowID, &s.TargetImage, &s.Status, &s.StartedAt, &completedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "Scan not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("DB query error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if completedAt.Valid {
		s.CompletedAt = &completedAt.String
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// GetScanReportHandler handles GET /scans/{id}/report
func (a *API) GetScanReportHandler(w http.ResponseWriter, r *http.Request) {
	scanID := r.PathValue("id")
	log.Printf("📥 Received report request for scan ID: %s", scanID)
	if scanID == "" {
		log.Printf("⚠️  Scan ID missing in request")
		http.Error(w, "scan id parameter is required", http.StatusBadRequest)
		return
	}

	query := `SELECT report_pdf FROM scans WHERE id = $1`

	var pdfBytes []byte
	err := a.DB.QueryRow(query, scanID).Scan(&pdfBytes)

	if err == sql.ErrNoRows {
		http.Error(w, "Scan not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("DB query error for report_pdf: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if len(pdfBytes) == 0 {
		log.Printf("⚠️  Report PDF is empty for scan ID: %s", scanID)
		http.Error(w, "Report PDF not yet generated or is empty", http.StatusNotFound)
		return
	}

	log.Printf("📤 Sending report PDF (%d bytes) for scan ID: %s", len(pdfBytes), scanID)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"scan_report_%s.pdf\"", scanID))
	w.Header().Set("Connection", "close")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	http.ServeContent(w, r, "report.pdf", time.Now(), bytes.NewReader(pdfBytes))

	log.Printf("✅ Drafted response for scan ID: %s", scanID)
}
