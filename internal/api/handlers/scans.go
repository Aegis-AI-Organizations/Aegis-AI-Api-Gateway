package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

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

	scanID, err := a.GRPCClient.StartScan(r.Context(), req.TargetImage)
	if err != nil {
		log.Printf("Failed to start scan via gRPC: %v", err)
		http.Error(w, "Failed to start workflow orchestrator", http.StatusInternalServerError)
		return
	}

	log.Printf("Started Orchestration Workflow for scanID: %s", scanID)

	res := models.CreateScanResponse{
		ScanID:             scanID,
		TemporalWorkflowID: fmt.Sprintf("pentest-workflow-%s", scanID),
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
	grpcScans, err := a.GRPCClient.ListScans(r.Context())
	if err != nil {
		log.Printf("GRPC query error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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

	grpcScans, err := a.GRPCClient.ListScans(r.Context())
	if err != nil {
		log.Printf("GRPC query error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var found *models.Scan
	for _, s := range grpcScans {
		if s.ScanId == scanID {
			startStr := ""
			if s.StartedAt != nil {
				startStr = s.StartedAt.AsTime().Format(time.RFC3339)
			}
			var compStr *string
			if s.CompletedAt != nil {
				t := s.CompletedAt.AsTime().Format(time.RFC3339)
				compStr = &t
			}
			found = &models.Scan{
				ID:                 s.ScanId,
				TemporalWorkflowID: s.TemporalWorkflowId,
				TargetImage:        s.TargetImage,
				Status:             s.Status,
				StartedAt:          startStr,
				CompletedAt:        compStr,
			}
			break
		}
	}

	if found == nil {
		http.Error(w, "Scan not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(found); err != nil {
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

	pdfBytes, err := a.GRPCClient.GetScanReport(r.Context(), scanID)
	if err != nil {
		// Just assuming any error means not found for simplicity since the grpc client returns error.
		log.Printf("GRPC GetScanReport error: %v", err)
		http.Error(w, "Scan or report not found", http.StatusNotFound)
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
