package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/models"
)

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
