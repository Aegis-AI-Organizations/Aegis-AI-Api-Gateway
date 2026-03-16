package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestNewRouterFull(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	mux := NewRouter(db, nil, nil)

	tests := []struct {
		method string
		path   string
		code   int
	}{
		{"GET", "/health", http.StatusOK},
		{"GET", "/", http.StatusOK},
		{"POST", "/scans", http.StatusBadRequest},
		{"GET", "/scans", http.StatusInternalServerError},
		{"GET", "/scans/1", http.StatusInternalServerError},
		{"GET", "/scans/1/vulnerabilities", http.StatusInternalServerError},
		{"GET", "/vulnerabilities/1/evidences", http.StatusInternalServerError},
		{"GET", "/scans/1/report", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		var body []byte
		if tt.method == "POST" {
			body = []byte(`{"target_image":"test"}`)
		}
		req, _ := http.NewRequest(tt.method, tt.path, bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		assert.NotEqual(t, http.StatusNotFound, rr.Code, "Path %s %s should be registered", tt.method, tt.path)
	}
}
