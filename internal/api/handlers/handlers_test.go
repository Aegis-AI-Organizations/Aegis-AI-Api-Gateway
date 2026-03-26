package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestCreateScanHandler_MethodNotAllowed(t *testing.T) {
	api := &API{}
	req, _ := http.NewRequest("GET", "/scans", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.CreateScanHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestCreateScanHandler_InvalidJSON(t *testing.T) {
	api := &API{}
	req, _ := http.NewRequest("POST", "/scans", bytes.NewBuffer([]byte("invalid-json")))
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.CreateScanHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateScanHandler_MissingImage(t *testing.T) {
	api := &API{}
	payload := map[string]string{"target_image": ""}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/scans", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.CreateScanHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetScanByIDHandler_MissingID(t *testing.T) {
	api := &API{}
	req, _ := http.NewRequest("GET", "/scans/", nil)
	// PathValue not set simulation
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetScanByIDHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetScanReportHandler_MissingID(t *testing.T) {
	api := &API{}
	req, _ := http.NewRequest("GET", "/scans//report", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetScanReportHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHealthHandler(t *testing.T) {
	api := &API{}
	req, _ := http.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.HealthHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "ok")
}

func TestRootHandler(t *testing.T) {
	api := &API{}
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.RootHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "aegis-api-gateway")
}
