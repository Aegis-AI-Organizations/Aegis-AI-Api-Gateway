package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetVulnerabilitiesHandler(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "vuln_type", "severity", "target_endpoint", "description", "discovered_at"}).
		AddRow("v1", "SQL Injection", "HIGH", "http://target", "Desc", now)

	mockDB.ExpectQuery("SELECT id, vuln_type, severity, target_endpoint, description, discovered_at FROM vulnerabilities WHERE scan_id =").
		WithArgs("s1").
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/scans/s1/vulnerabilities", nil)
	req.SetPathValue("id", "s1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetVulnerabilitiesHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetEvidencesHandler(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "vulnerability_id", "payload_used", "loot_data", "captured_at"}).
		AddRow("e1", "v1", "payload", "loot", now)

	mockDB.ExpectQuery("SELECT id, vulnerability_id, payload_used, loot_data, captured_at FROM evidences WHERE vulnerability_id =").
		WithArgs("v1").
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/vulnerabilities/v1/evidences", nil)
	req.SetPathValue("id", "v1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetEvidencesHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetVulnerabilitiesHandler_DBError(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}
	mockDB.ExpectQuery("SELECT id, vuln_type, severity, target_endpoint, description, discovered_at FROM vulnerabilities WHERE scan_id =").
		WillReturnError(fmt.Errorf("db error"))

	req, _ := http.NewRequest("GET", "/scans/s1/vulnerabilities", nil)
	req.SetPathValue("id", "s1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetVulnerabilitiesHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestGetEvidencesHandler_DBError(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}
	mockDB.ExpectQuery("SELECT id, vulnerability_id, payload_used, loot_data, captured_at FROM evidences WHERE vulnerability_id =").
		WillReturnError(fmt.Errorf("db error"))

	req, _ := http.NewRequest("GET", "/vulnerabilities/v1/evidences", nil)
	req.SetPathValue("id", "v1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetEvidencesHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestGetVulnerabilitiesHandler_RowErr(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}
	rows := sqlmock.NewRows([]string{"id", "vuln_type", "severity", "target_endpoint", "description", "discovered_at"}).
		AddRow("v1", "type", "high", "end", "desc", time.Now()).
		RowError(0, fmt.Errorf("row error"))

	mockDB.ExpectQuery("SELECT id, vuln_type, severity, target_endpoint, description, discovered_at FROM vulnerabilities WHERE scan_id =").
		WithArgs("s1").
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/scans/s1/vulnerabilities", nil)
	req.SetPathValue("id", "s1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetVulnerabilitiesHandler)
	handler.ServeHTTP(rr, req)
}

func TestGetVulnerabilitiesHandler_Empty(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}
	mockDB.ExpectQuery("SELECT id, vuln_type, severity, target_endpoint, description, discovered_at FROM vulnerabilities WHERE scan_id =").
		WillReturnRows(sqlmock.NewRows([]string{"id", "vuln_type", "severity", "target_endpoint", "description", "discovered_at"}))

	req, _ := http.NewRequest("GET", "/scans/s1/vulnerabilities", nil)
	req.SetPathValue("id", "s1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetVulnerabilitiesHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetEvidencesHandler_Empty(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}
	mockDB.ExpectQuery("SELECT id, vulnerability_id, payload_used, loot_data, captured_at FROM evidences WHERE vulnerability_id =").
		WillReturnRows(sqlmock.NewRows([]string{"id", "vulnerability_id", "payload_used", "loot_data", "captured_at"}))

	req, _ := http.NewRequest("GET", "/vulnerabilities/v1/evidences", nil)
	req.SetPathValue("id", "v1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetEvidencesHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetEvidencesHandler_RowScanError(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}
	// Wrong type for id
	rows := sqlmock.NewRows([]string{"id", "vulnerability_id", "payload_used", "loot_data", "captured_at"}).
		AddRow(1, "v1", "payload", "loot", time.Now())

	mockDB.ExpectQuery("SELECT id, vulnerability_id, payload_used, loot_data, captured_at FROM evidences WHERE vulnerability_id =").
		WithArgs("v1").
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/vulnerabilities/v1/evidences", nil)
	req.SetPathValue("id", "v1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetEvidencesHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetEvidencesHandler_RowErr(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}
	rows := sqlmock.NewRows([]string{"id", "vulnerability_id", "payload_used", "loot_data", "captured_at"}).
		AddRow("e1", "v1", "payload", "loot", time.Now()).
		RowError(0, fmt.Errorf("row error"))

	mockDB.ExpectQuery("SELECT id, vulnerability_id, payload_used, loot_data, captured_at FROM evidences WHERE vulnerability_id =").
		WithArgs("v1").
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/vulnerabilities/v1/evidences", nil)
	req.SetPathValue("id", "v1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetEvidencesHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
