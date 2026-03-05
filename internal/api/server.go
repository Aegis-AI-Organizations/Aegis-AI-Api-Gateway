package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/middleware"
	"go.temporal.io/sdk/client"
)

func NewRouter(db *sql.DB, tc client.Client) *http.ServeMux {
	mux := http.NewServeMux()

	h := &handlers.API{
		DB:             db,
		TemporalClient: tc,
	}

	mux.HandleFunc("GET /health", h.HealthHandler)
	mux.HandleFunc("GET /", h.RootHandler)

	mux.HandleFunc("POST /scans", h.CreateScanHandler)
	mux.HandleFunc("GET /scans", h.GetScansHandler)
	mux.HandleFunc("GET /scans/{id}", h.GetScanByIDHandler)

	mux.HandleFunc("GET /scans/{id}/vulnerabilities", h.GetVulnerabilitiesHandler)
	mux.HandleFunc("GET /vulnerabilities/{id}/evidences", h.GetEvidencesHandler)

	return mux
}

func Start(database *sql.DB, tc client.Client) {
	fmt.Println("🌍 Aegis AI Web API Gateway HTTP Server starting...")

	router := NewRouter(database, tc)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🌍 Listening on :%s", port)
	if err := http.ListenAndServe(":"+port, middleware.CORS(router)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
