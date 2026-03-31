package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc/aegis/v2"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/testutils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

type MockScanStream struct {
	mock.Mock
}

func (m *MockScanStream) Recv() (*v1.WatchScanStatusResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.WatchScanStatusResponse), args.Error(1)
}

func (m *MockScanStream) Context() context.Context { return context.Background() }
func (m *MockScanStream) Header() (metadata.MD, error) { return nil, nil }
func (m *MockScanStream) Trailer() metadata.MD { return nil }
func (m *MockScanStream) CloseSend() error { return nil }
func (m *MockScanStream) SendMsg(m_ interface{}) error { return nil }
func (m *MockScanStream) RecvMsg(m_ interface{}) error { return nil }


func TestScanStreamHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockScanServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}

	mockStream := new(MockScanStream)
	mockStream.On("Recv").Return(&v1.WatchScanStatusResponse{ScanId: "s1", Status: "RUNNING"}, nil).Once()
	mockStream.On("Recv").Return(nil, fmt.Errorf("EOF")).Once()

	mockService.On("WatchScanStatus", mock.Anything, &v1.WatchScanStatusRequest{ScanId: "s1"}).
		Return(mockStream, nil)

	w := testutils.NewCloseNotifierRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/scans/s1/stream", nil)
	c.Params = []gin.Param{{Key: "id", Value: "s1"}}

	api.ScanStreamHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "data:{\"scan_id\":\"s1\",\"status\":\"RUNNING\"}")
}

func TestScanStreamHandler_Global(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockScanServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}

	mockStream := new(MockScanStream)
	mockStream.On("Recv").Return(&v1.WatchScanStatusResponse{ScanId: "s2", Status: "COMPLETED"}, nil).Once()
	mockStream.On("Recv").Return(nil, fmt.Errorf("EOF")).Once()

	mockService.On("WatchScanStatus", mock.Anything, &v1.WatchScanStatusRequest{ScanId: ""}).
		Return(mockStream, nil)

	w := testutils.NewCloseNotifierRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/scans/stream", nil)

	api.ScanStreamHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "data:{\"scan_id\":\"s2\",\"status\":\"COMPLETED\"}")
}

func TestScanStreamHandler_GRPCError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockScanServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			ScanService: mockService,
		},
	}

	mockService.On("WatchScanStatus", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("grpc error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/scans/s1/stream", nil)
	c.Params = []gin.Param{{Key: "id", Value: "s1"}}

	api.ScanStreamHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
