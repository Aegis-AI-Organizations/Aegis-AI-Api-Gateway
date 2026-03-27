package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// ScanStreamHandler handles GET /scans/stream or /scans/{id}/stream
func (a *API) ScanStreamHandler(w http.ResponseWriter, r *http.Request) {
	scanID := r.PathValue("id")
	if scanID == "" {
		log.Printf("📡 Starting global SSE stream")
	} else {
		log.Printf("📡 Starting SSE stream for scanID: %s", scanID)
	}

	stream, err := a.GRPCClient.WatchScanStatus(r.Context(), scanID)
	if err != nil || stream == nil {
		log.Printf("Failed to open gRPC stream: %v", err)
		http.Error(w, "Failed to initialize stream", http.StatusInternalServerError)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			log.Printf("gRPC stream closed: %v", err)
			return
		}

		payload, err := json.Marshal(map[string]string{
			"scan_id": resp.ScanId,
			"status":  resp.Status,
		})
		if err != nil {
			log.Printf("Failed to marshal SSE payload: %v", err)
			continue
		}

		if _, err := fmt.Fprintf(w, "data: %s\n\n", string(payload)); err != nil {
			log.Printf("Failed to write to SSE stream: %v", err)
			return
		}
		flusher.Flush()
	}
}
