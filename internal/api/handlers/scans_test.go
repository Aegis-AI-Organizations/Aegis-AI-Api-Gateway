package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/client"
)

type MockWorkflowRun struct {
	mock.Mock
}

func (m *MockWorkflowRun) GetID() string {
	return "test-workflow-id"
}

func (m *MockWorkflowRun) GetRunID() string {
	return "test-run-id"
}

func (m *MockWorkflowRun) Get(ctx context.Context, valuePtr interface{}) error {
	return nil
}

func (m *MockWorkflowRun) GetWithOptions(ctx context.Context, valuePtr interface{}, options client.WorkflowRunGetOptions) error {
	return nil
}

type MockTemporalClient struct {
	client.Client
	mock.Mock
}

func (m *MockTemporalClient) ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	callArgs := m.Called(ctx, options, workflow, args)
	return callArgs.Get(0).(client.WorkflowRun), callArgs.Error(1)
}

func TestCreateScanHandler(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	mockTemporal := new(MockTemporalClient)
	api := &API{
		DB:             db,
		TemporalClient: mockTemporal,
	}

	payload := map[string]string{"target_image": "nginx:latest"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/scans", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	mockDB.ExpectExec("INSERT INTO scans").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "nginx:latest").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mockRun := new(MockWorkflowRun)
	mockTemporal.On("ExecuteWorkflow", mock.Anything, mock.Anything, "PentestWorkflow", mock.Anything).
		Return(mockRun, nil)

	handler := http.HandlerFunc(api.CreateScanHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockTemporal.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestCreateScanHandler_DBFailure(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}

	payload := map[string]string{"target_image": "nginx:latest"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/scans", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	mockDB.ExpectExec("INSERT INTO scans").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "nginx:latest").
		WillReturnError(fmt.Errorf("db error"))

	handler := http.HandlerFunc(api.CreateScanHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestCreateScanHandler_TemporalFailure(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	mockTemporal := new(MockTemporalClient)
	api := &API{
		DB:             db,
		TemporalClient: mockTemporal,
	}

	payload := map[string]string{"target_image": "nginx:latest"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/scans", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	mockDB.ExpectExec("INSERT INTO scans").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "nginx:latest").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mockTemporal.On("ExecuteWorkflow", mock.Anything, mock.Anything, "PentestWorkflow", mock.Anything).
		Return((*MockWorkflowRun)(nil), fmt.Errorf("temporal error"))

	mockDB.ExpectExec("UPDATE scans SET status = 'FAILED'").
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	handler := http.HandlerFunc(api.CreateScanHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestGetScansHandler_DBError(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}
	mockDB.ExpectQuery("SELECT id, temporal_workflow_id, target_image, status, started_at, completed_at FROM scans").
		WillReturnError(fmt.Errorf("db error"))

	req, _ := http.NewRequest("GET", "/scans", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetScansHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestGetScanByIDHandler_NotFound(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}
	mockDB.ExpectQuery("SELECT id, temporal_workflow_id, target_image, status, started_at, completed_at FROM scans WHERE id =").
		WithArgs("nb").
		WillReturnError(sql.ErrNoRows)

	req, _ := http.NewRequest("GET", "/scans/nb", nil)
	req.SetPathValue("id", "nb")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetScanByIDHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetScanReportHandler_EmptyReport(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}
	rows := sqlmock.NewRows([]string{"report_pdf"}).AddRow([]byte{})

	mockDB.ExpectQuery("SELECT report_pdf FROM scans WHERE id =").
		WithArgs("s1").
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/scans/s1/report", nil)
	req.SetPathValue("id", "s1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetScanReportHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetScanReportHandler_DBError(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}
	mockDB.ExpectQuery("SELECT report_pdf FROM scans WHERE id =").
		WithArgs("s1").
		WillReturnError(fmt.Errorf("db error"))

	req, _ := http.NewRequest("GET", "/scans/s1/report", nil)
	req.SetPathValue("id", "s1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetScanReportHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestGetScanReportHandler_NotFound(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}
	mockDB.ExpectQuery("SELECT report_pdf FROM scans WHERE id =").
		WithArgs("s1").
		WillReturnError(sql.ErrNoRows)

	req, _ := http.NewRequest("GET", "/scans/s1/report", nil)
	req.SetPathValue("id", "s1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetScanReportHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetScansHandler_RowScanError(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}
	// Add a row with a wrong type to trigger scan error
	rows := sqlmock.NewRows([]string{"id", "temporal_workflow_id", "target_image", "status", "started_at", "completed_at"}).
		AddRow(1, "wf-1", "img-1", "PENDING", "start", "end")

	mockDB.ExpectQuery("SELECT id, temporal_workflow_id, target_image, status, started_at, completed_at FROM scans").
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/scans", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetScansHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetScanByIDHandler_DBError(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}
	mockDB.ExpectQuery("SELECT id, temporal_workflow_id, target_image, status, started_at, completed_at FROM scans WHERE id =").
		WithArgs("s1").
		WillReturnError(fmt.Errorf("db error"))

	req, _ := http.NewRequest("GET", "/scans/s1", nil)
	req.SetPathValue("id", "s1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetScanByIDHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestGetScanReportHandler_Success(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}
	pdfData := []byte("fake-pdf-content")
	rows := sqlmock.NewRows([]string{"report_pdf"}).AddRow(pdfData)

	mockDB.ExpectQuery("SELECT report_pdf FROM scans WHERE id =").
		WithArgs("s1").
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/scans/s1/report", nil)
	req.SetPathValue("id", "s1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetScanReportHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/pdf", rr.Header().Get("Content-Type"))
}

func TestGetScansHandler_RowErr(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	api := &API{DB: db}
	rows := sqlmock.NewRows([]string{"id", "temporal_workflow_id", "target_image", "status", "started_at", "completed_at"}).
		AddRow("s1", "wf1", "img1", "PENDING", time.Now(), nil).
		RowError(0, fmt.Errorf("row error"))

	mockDB.ExpectQuery("SELECT id, temporal_workflow_id, target_image, status, started_at, completed_at FROM scans").
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/scans", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(api.GetScansHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
