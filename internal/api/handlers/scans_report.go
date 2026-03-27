package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			log.Printf("GRPC GetScanReport not found: %v", err)
			http.Error(w, "Scan or report not found", http.StatusNotFound)
		} else {
			log.Printf("GRPC GetScanReport error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
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
