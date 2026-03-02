package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
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

	return mux
}

func Start(database *sql.DB, tc client.Client) {
	fmt.Println("🌍 Aegis AI Web API Gateway HTTP Server starting...")

	router := NewRouter(database, tc)

	log.Println("🌍 Listening on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
