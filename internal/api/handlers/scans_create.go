package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/models"
)

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

	resp, err := a.GRPCClient.StartScan(r.Context(), req.TargetImage)
	if err != nil {
		log.Printf("Failed to start scan via gRPC: %v", err)
		http.Error(w, "Failed to start workflow orchestrator", http.StatusInternalServerError)
		return
	}

	log.Printf("Started Orchestration Workflow for scanID: %s", resp.ScanId)

	res := models.CreateScanResponse{
		ScanID: resp.ScanId,
		Status: resp.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
