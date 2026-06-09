package handlers_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type MockTopologyService struct {
	mock.Mock
}

func (m *MockTopologyService) GetTopology(ctx context.Context, companyID string) (handlers.TopologyResponse, error) {
	args := m.Called(ctx, companyID)
	if args.Get(0) == nil {
		return handlers.TopologyResponse{}, args.Error(1)
	}
	return args.Get(0).(handlers.TopologyResponse), args.Error(1)
}

func TestGetTopologyHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockTopology := new(MockTopologyService)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{},
		Topology:   mockTopology,
	}

	expected := handlers.TopologyResponse{
		Hosts: []handlers.TopologyHost{
			{
				ID:       "api-route",
				Hostname: "api.aegis-ai.fr",
			},
		},
	}

	mockTopology.On("GetTopology", mock.Anything, "company-123").Return(expected, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("company_id", "company-123")
	c.Request, _ = http.NewRequest("GET", "/api/topology", nil)

	api.GetTopologyHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"api.aegis-ai.fr"`)
	mockTopology.AssertExpectations(t)
}

func TestNeo4jTopologyService_AppendsPublicApiRoute(t *testing.T) {
	callCount := 0
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			callCount++
			assert.Equal(t, http.MethodPost, req.Method)
			assert.Contains(t, req.URL.Path, "/db/neo4j/tx/commit")
			assert.Contains(t, req.Header.Get("Authorization"), "Basic ")

			body := io.NopCloser(strings.NewReader(`{"results":[{"data":[]}],"errors":[]}`))
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       body,
				Request:    req,
			}, nil
		}),
	}

	service := handlers.NewNeo4jTopologyService("http://neo4j.local:7474", "neo4j", "secret", "neo4j", client)
	resp, err := service.GetTopology(context.Background(), "company-123")

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, callCount, 1)
	if assert.Len(t, resp.Hosts, 1) {
		assert.Equal(t, "api-route", resp.Hosts[0].ID)
		assert.Equal(t, "api.aegis-ai.fr", resp.Hosts[0].Hostname)
	}
}
