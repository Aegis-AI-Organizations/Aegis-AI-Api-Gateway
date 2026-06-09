package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMinioClient struct {
	mock.Mock
}

func (m *MockMinioClient) GeneratePresignedPutURL(ctx context.Context, objectName string) (string, error) {
	args := m.Called(ctx, objectName)
	return args.String(0), args.Error(1)
}

func TestGetUploadURLHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockMinio := new(MockMinioClient)
	handler := &handlers.MinioHandler{MinioClient: mockMinio}

	mockMinio.On("GeneratePresignedPutURL", mock.Anything, mock.Anything).
		Return("http://minio/upload/url", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/storage/upload-url?prefix=test", nil)
	c.Set("company_id", "c1")

	handler.GetUploadURLHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "http://minio/upload/url")
}

func TestGetUploadURLHandler_NoClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &handlers.MinioHandler{MinioClient: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/storage/upload-url", nil)

	handler.GetUploadURLHandler(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
