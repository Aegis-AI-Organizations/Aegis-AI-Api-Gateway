package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc/aegis/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MockVulnerabilityServiceClient struct {
	mock.Mock
}

func (m *MockVulnerabilityServiceClient) GetVulnerabilities(ctx context.Context, in *v1.GetVulnerabilitiesRequest, opts ...grpc.CallOption) (*v1.GetVulnerabilitiesResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.GetVulnerabilitiesResponse), args.Error(1)
}

func (m *MockVulnerabilityServiceClient) GetEvidences(ctx context.Context, in *v1.GetEvidencesRequest, opts ...grpc.CallOption) (*v1.GetEvidencesResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.GetEvidencesResponse), args.Error(1)
}

func TestGetVulnerabilitiesHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockVulnerabilityServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			VulnerabilityService: mockService,
		},
	}

	response := &v1.GetVulnerabilitiesResponse{
		Vulnerabilities: []*v1.Vulnerability{
			{
				Id:             "v1",
				VulnType:       "SQL Injection",
				Severity:       "HIGH",
				TargetEndpoint: "http://target",
				Description:    "Desc",
				DiscoveredAt:   timestamppb.New(time.Now()),
			},
		},
	}

	mockService.On("GetVulnerabilities", mock.Anything, &v1.GetVulnerabilitiesRequest{ScanId: "s1"}).Return(response, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/scans/s1/vulnerabilities", nil)
	c.Params = []gin.Param{{Key: "id", Value: "s1"}}

	api.GetVulnerabilitiesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetEvidencesHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockVulnerabilityServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			VulnerabilityService: mockService,
		},
	}

	response := &v1.GetEvidencesResponse{
		Evidences: []*v1.Evidence{
			{
				Id:              "e1",
				VulnerabilityId: "v1",
				PayloadUsed:     "payload",
				LootData:        "loot",
				CapturedAt:      timestamppb.New(time.Now()),
			},
		},
	}

	mockService.On("GetEvidences", mock.Anything, &v1.GetEvidencesRequest{VulnerabilityId: "v1"}).Return(response, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/vulnerabilities/v1/evidences", nil)
	c.Params = []gin.Param{{Key: "id", Value: "v1"}}

	api.GetEvidencesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetVulnerabilitiesHandler_GRPCError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockVulnerabilityServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			VulnerabilityService: mockService,
		},
	}

	mockService.On("GetVulnerabilities", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("grpc error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/scans/s1/vulnerabilities", nil)
	c.Params = []gin.Param{{Key: "id", Value: "s1"}}

	api.GetVulnerabilitiesHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetEvidencesHandler_GRPCError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockVulnerabilityServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			VulnerabilityService: mockService,
		},
	}

	mockService.On("GetEvidences", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("grpc error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/vulnerabilities/v1/evidences", nil)
	c.Params = []gin.Param{{Key: "id", Value: "v1"}}

	api.GetEvidencesHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetVulnerabilitiesHandler_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockVulnerabilityServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			VulnerabilityService: mockService,
		},
	}

	response := &v1.GetVulnerabilitiesResponse{
		Vulnerabilities: []*v1.Vulnerability{},
	}

	mockService.On("GetVulnerabilities", mock.Anything, mock.Anything).Return(response, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/scans/s1/vulnerabilities", nil)
	c.Params = []gin.Param{{Key: "id", Value: "s1"}}

	api.GetVulnerabilitiesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetEvidencesHandler_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockVulnerabilityServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			VulnerabilityService: mockService,
		},
	}

	response := &v1.GetEvidencesResponse{
		Evidences: []*v1.Evidence{},
	}

	mockService.On("GetEvidences", mock.Anything, mock.Anything).Return(response, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/vulnerabilities/v1/evidences", nil)
	c.Params = []gin.Param{{Key: "id", Value: "v1"}}

	api.GetEvidencesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
