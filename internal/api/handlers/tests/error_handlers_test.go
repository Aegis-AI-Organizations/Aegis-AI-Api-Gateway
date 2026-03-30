package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
)

type errorResponseWriter struct {
	http.ResponseWriter
}

func (e *errorResponseWriter) Write(b []byte) (int, error) {
	return 0, errors.New("write error")
}

func (e *errorResponseWriter) Header() http.Header {
	return http.Header{}
}

func (e *errorResponseWriter) WriteHeader(statusCode int) {}

func TestHandlers_WriteError(t *testing.T) {
	api := &handlers.API{}
	req := httptest.NewRequest("GET", "/health", nil)

	// Test HealthHandler error path
	rrHealth := &errorResponseWriter{}
	api.HealthHandler(rrHealth, req)

	// Test RootHandler error path
	rrRoot := &errorResponseWriter{}
	api.RootHandler(rrRoot, req)
}
