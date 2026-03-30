package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
	"github.com/stretchr/testify/assert"
)

type errorResponseWriter struct {
	http.ResponseWriter
	WriteCalled bool
}

func (e *errorResponseWriter) Write(b []byte) (int, error) {
	e.WriteCalled = true
	return 0, errors.New("write error")
}

func (e *errorResponseWriter) Header() http.Header {
	return http.Header{}
}

func (e *errorResponseWriter) WriteHeader(statusCode int) {}

func TestHandlers_WriteError(t *testing.T) {
	api := &handlers.API{}
	req := httptest.NewRequest("GET", "/health", nil)

	t.Run("HealthHandler", func(t *testing.T) {
		rr := &errorResponseWriter{}
		assert.NotPanics(t, func() {
			api.HealthHandler(rr, req)
		})
		assert.True(t, rr.WriteCalled, "Write should have been called on HealthHandler")
	})

	t.Run("RootHandler", func(t *testing.T) {
		rr := &errorResponseWriter{}
		assert.NotPanics(t, func() {
			api.RootHandler(rr, req)
		})
		assert.True(t, rr.WriteCalled, "Write should have been called on RootHandler")
	})
}
