package handlers

import (
	"fmt"
	"log"
	"net/http"
)

// ScanStreamHandler handles GET /scans/stream or /scans/{id}/stream
func (a *API) ScanStreamHandler(w http.ResponseWriter, r *http.Request) {
	scanID := r.PathValue("id")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	log.Printf("📡 Starting SSE stream for scanID: %s", scanID)

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
			log.Printf("gRPC stream closed for scanID %s: %v", scanID, err)
			return
		}

		if _, err := fmt.Fprintf(w, "data: {\"scan_id\": \"%s\", \"status\": \"%s\"}\n\n", resp.ScanId, resp.Status); err != nil {
			log.Printf("Failed to write to SSE stream: %v", err)
			return
		}
		flusher.Flush()
	}
}
