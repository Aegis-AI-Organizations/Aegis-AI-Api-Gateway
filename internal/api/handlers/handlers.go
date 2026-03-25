package handlers

import (
	"log"
	"net/http"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/grpc"
)

// API holds the core dependencies dynamically injected by our server initialization.
type API struct {
	GRPCClient     *grpc.Client
}

// HealthHandler returns a simple 200 OK status for Kubernetes liveness probes.
func (a *API) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// RootHandler returns the service name and version.
func (a *API) RootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"service":"aegis-api-gateway","version":"pre-alpha"}`)); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
