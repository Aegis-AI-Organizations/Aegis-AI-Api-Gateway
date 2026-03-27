package api

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/middleware"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/grpc"
)

func NewRouter(gc *grpc.Client) *http.ServeMux {
	mux := http.NewServeMux()

	h := &handlers.API{
		GRPCClient:     gc,
	}

	mux.HandleFunc("GET /health", h.HealthHandler)
	mux.HandleFunc("GET /", h.RootHandler)

	mux.HandleFunc("POST /scans", h.CreateScanHandler)
	mux.HandleFunc("GET /scans", h.GetScansHandler)
	mux.HandleFunc("GET /scans/{id}", h.GetScanByIDHandler)

	mux.HandleFunc("GET /scans/{id}/vulnerabilities", h.GetVulnerabilitiesHandler)
	mux.HandleFunc("GET /vulnerabilities/{id}/evidences", h.GetEvidencesHandler)

	mux.HandleFunc("GET /scans/{id}/report", h.GetScanReportHandler)

	mux.HandleFunc("GET /scans/stream", h.ScanStreamHandler)
	mux.HandleFunc("GET /scans/{id}/stream", h.ScanStreamHandler)

	return mux
}

func Start(gc *grpc.Client) {
	fmt.Println("🌍 Aegis AI Web API Gateway HTTP Server starting...")

	router := NewRouter(gc)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🌍 Listening on :%s", port)
	if err := http.ListenAndServe(":"+port, middleware.CORS(router)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
