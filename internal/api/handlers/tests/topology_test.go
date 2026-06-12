package handlers_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/types"
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

func (m *MockTopologyService) GetTopologyDebug(ctx context.Context, companyID string) (handlers.TopologyDebugResponse, error) {
	args := m.Called(ctx, companyID)
	if args.Get(0) == nil {
		return handlers.TopologyDebugResponse{}, args.Error(1)
	}
	return args.Get(0).(handlers.TopologyDebugResponse), args.Error(1)
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
				ID:       "host-1",
				Hostname: "host-a.local",
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
	assert.Contains(t, w.Body.String(), `"host-1"`)
	mockTopology.AssertExpectations(t)
}

func TestGetTopologyHandler_AllowsAdminCompanyOverride(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockTopology := new(MockTopologyService)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{},
		Topology:   mockTopology,
	}

	expected := handlers.TopologyResponse{
		Hosts: []handlers.TopologyHost{{ID: "host-override", Hostname: "company-b.local"}},
	}

	mockTopology.On("GetTopology", mock.Anything, "company-456").Return(expected, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(string(types.CompanyIDKey), "company-123")
	c.Set(string(types.RoleKey), string(types.RoleAdmin))
	c.Request, _ = http.NewRequest("GET", "/api/topology?company_id=company-456", nil)

	api.GetTopologyHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"company-b.local"`)
	mockTopology.AssertExpectations(t)
}

func TestGetTopologyHandler_AllowsAdminGlobalScope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockTopology := new(MockTopologyService)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{},
		Topology:   mockTopology,
	}

	expected := handlers.TopologyResponse{
		Hosts: []handlers.TopologyHost{{ID: "host-global", Hostname: "all-companies.local"}},
	}

	mockTopology.On("GetTopology", mock.Anything, "all").Return(expected, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(string(types.CompanyIDKey), "company-123")
	c.Set(string(types.RoleKey), string(types.RoleAdmin))
	c.Request, _ = http.NewRequest("GET", "/api/topology?company_id=all", nil)

	api.GetTopologyHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"all-companies.local"`)
	mockTopology.AssertExpectations(t)
}

func TestGetTopologyDebugHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockTopology := new(MockTopologyService)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{},
		Topology:   mockTopology,
	}

	expected := handlers.TopologyDebugResponse{
		CompanyID: "company-123",
		Summary: handlers.TopologyDebugSummary{
			Hosts:      1,
			Containers: 1,
			Relations:  2,
		},
		Agents: []string{"agent-1"},
		Hosts: []handlers.TopologyHost{
			{
				ID: "host-1",
				Containers: []handlers.TopologyContainer{
					{ID: "container-1", Name: "api"},
				},
			},
		},
		Relations: []handlers.TopologyDebugRelation{
			{Type: "RUNS_CONTAINER", Source: "host-1", Target: "api"},
			{Type: "DEPENDS_ON", Source: "api", Target: "postgres"},
		},
	}

	mockTopology.On("GetTopologyDebug", mock.Anything, "company-123").Return(expected, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(string(types.CompanyIDKey), "company-123")
	c.Request, _ = http.NewRequest("GET", "/api/admin/topology/latest", nil)

	api.GetTopologyDebugHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"company_id":"company-123"`)
	assert.Contains(t, w.Body.String(), `"agents":["agent-1"]`)
	assert.Contains(t, w.Body.String(), `"type":"DEPENDS_ON"`)
	mockTopology.AssertExpectations(t)
}

func TestNeo4jTopologyService_DoesNotAppendPublicApiRoute(t *testing.T) {
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
	assert.Empty(t, resp.Hosts)
}

func TestNeo4jTopologyService_RendersNeo4jTopologyGraph(t *testing.T) {
	callCount := 0
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			callCount++
			body, _ := io.ReadAll(req.Body)
			var payload struct {
				Statements []struct {
					Statement string `json:"statement"`
				} `json:"statements"`
			}
			_ = json.Unmarshal(body, &payload)
			statement := ""
			if len(payload.Statements) > 0 {
				statement = payload.Statements[0].Statement
			}

			var responseBody string
			switch {
			case strings.Contains(statement, "MATCH (h:Host)-[:RUNS_CONTAINER]->(c:Container)-[:RUNS_PROCESS]->(p:Process)"):
				responseBody = `{"results":[{"data":[{"row":["host-1","container-1","process-2",42,"nginx","nginx -g daemon off;","www-data"]}]}],"errors":[]}`
			case strings.Contains(statement, "MATCH (h:Host)-[:RUNS_PROCESS]->(p:Process)"):
				responseBody = `{"results":[{"data":[{"row":["host-1","process-1",12,"sshd","/usr/sbin/sshd","root"]}]}],"errors":[]}`
			case strings.Contains(statement, "MATCH (h:Host)-[:RUNS_CONTAINER]->(c:Container)"):
				responseBody = `{"results":[{"data":[{"row":["host-1","host-1.local",["10.0.0.1"],"container-1","web","nginx:latest",{"FOO":"bar"},{"com.docker.compose.service":"web"},["frontend"],["80:tcp:LISTEN"],["8080:tcp:LISTEN"]]}]}],"errors":[]}`
			case strings.Contains(statement, "MATCH (h:Host)"):
				responseBody = `{"results":[{"data":[{"row":["host-1","host-1.local",["10.0.0.1"]]}]}],"errors":[]}`
			default:
				responseBody = `{"results":[{"data":[]}],"errors":[]}`
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(responseBody)),
				Request:    req,
			}, nil
		}),
	}

	service := handlers.NewNeo4jTopologyService("http://neo4j.local:7474", "neo4j", "secret", "neo4j", client)
	resp, err := service.GetTopology(context.Background(), "company-123")

	assert.NoError(t, err)
	assert.Equal(t, 4, callCount)
	if assert.Len(t, resp.Hosts, 1) {
		host := resp.Hosts[0]
		assert.Equal(t, "host-1", host.ID)
		assert.Equal(t, []string{"10.0.0.1"}, host.IPAddresses)
		assert.Len(t, host.Containers, 1)
		assert.Len(t, host.Processes, 1)
		assert.Equal(t, "web", host.Containers[0].Name)
		assert.Equal(t, "nginx:latest", host.Containers[0].Image)
		assert.Equal(t, map[string]string{"FOO": "bar"}, host.Containers[0].Env)
		assert.Equal(t, map[string]string{"com.docker.compose.service": "web"}, host.Containers[0].Labels)
		assert.Equal(t, []string{"frontend"}, host.Containers[0].Networks)
		assert.Len(t, host.Containers[0].Ports, 1)
		assert.Len(t, host.Containers[0].ExposedPorts, 1)
		assert.Len(t, host.Containers[0].Processes, 1)
	}
}

func TestNeo4jTopologyService_RendersTopologyDebug(t *testing.T) {
	callCount := 0
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			callCount++
			body, _ := io.ReadAll(req.Body)
			var payload struct {
				Statements []struct {
					Statement string `json:"statement"`
				} `json:"statements"`
			}
			_ = json.Unmarshal(body, &payload)
			statement := ""
			if len(payload.Statements) > 0 {
				statement = payload.Statements[0].Statement
			}

			var responseBody string
			switch {
			case strings.Contains(statement, "MATCH (source)-[rel:RUNS_CONTAINER|RUNS_PROCESS|CONNECTED_TO|DEPENDS_ON|ROUTE_FROM|ROUTE_TO]->(target)"):
				responseBody = `{"results":[{"data":[{"row":["RUNS_CONTAINER","host-1","api"]},{"row":["DEPENDS_ON","api","postgres"]}]}],"errors":[]}`
			case strings.Contains(statement, "WITH DISTINCT n.agentId AS agent_id"):
				responseBody = `{"results":[{"data":[{"row":["agent-1"]},{"row":["agent-1"]},{"row":["agent-2"]}]}],"errors":[]}`
			case strings.Contains(statement, "MATCH (h:Host)-[:RUNS_CONTAINER]->(c:Container)-[:RUNS_PROCESS]->(p:Process)"):
				responseBody = `{"results":[{"data":[{"row":["host-1","container-1","process-1",7,"node","node server.js","node"]}]}],"errors":[]}`
			case strings.Contains(statement, "MATCH (h:Host)-[:RUNS_PROCESS]->(p:Process)"):
				responseBody = `{"results":[{"data":[{"row":["host-1","process-2",1,"systemd","/sbin/init","root"]}]}],"errors":[]}`
			case strings.Contains(statement, "MATCH (h:Host)-[:RUNS_CONTAINER]->(c:Container)"):
				responseBody = `{"results":[{"data":[{"row":["host-1","host-1.local",["10.0.0.1"],"container-1","api","node:20",["DB_HOST=postgres"],["com.docker.compose.service=api"],["backend"],["3000:tcp:LISTEN"],["3000:tcp:LISTEN"]]}]}],"errors":[]}`
			case strings.Contains(statement, "MATCH (h:Host)"):
				responseBody = `{"results":[{"data":[{"row":["host-1","host-1.local",["10.0.0.1"]]}]}],"errors":[]}`
			default:
				responseBody = `{"results":[{"data":[]}],"errors":[]}`
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(responseBody)),
				Request:    req,
			}, nil
		}),
	}

	service := handlers.NewNeo4jTopologyService("http://neo4j.local:7474", "neo4j", "secret", "neo4j", client)
	resp, err := service.GetTopologyDebug(context.Background(), "company-123")

	assert.NoError(t, err)
	assert.Equal(t, 6, callCount)
	assert.Equal(t, "company-123", resp.CompanyID)
	assert.Equal(t, []string{"agent-1", "agent-2"}, resp.Agents)
	assert.Equal(t, handlers.TopologyDebugSummary{
		Hosts:      1,
		Containers: 1,
		Processes:  2,
		Relations:  2,
	}, resp.Summary)
	if assert.Len(t, resp.Relations, 2) {
		assert.Equal(t, handlers.TopologyDebugRelation{Type: "RUNS_CONTAINER", Source: "host-1", Target: "api"}, resp.Relations[0])
		assert.Equal(t, handlers.TopologyDebugRelation{Type: "DEPENDS_ON", Source: "api", Target: "postgres"}, resp.Relations[1])
	}
	if assert.Len(t, resp.Hosts, 1) && assert.Len(t, resp.Hosts[0].Containers, 1) {
		container := resp.Hosts[0].Containers[0]
		assert.Equal(t, map[string]string{"DB_HOST": "postgres"}, container.Env)
		assert.Equal(t, map[string]string{"com.docker.compose.service": "api"}, container.Labels)
		assert.Equal(t, []string{"backend"}, container.Networks)
	}
}

func TestNeo4jTopologyService_RejectsEmptyCompany(t *testing.T) {
	service := handlers.NewNeo4jTopologyService("http://neo4j.local:7474", "neo4j", "secret", "neo4j", &http.Client{})
	_, err := service.GetTopology(context.Background(), " ")

	assert.Error(t, err)
}

func TestNeo4jTopologyService_AllScopeOmitsCompanyFilter(t *testing.T) {
	callCount := 0
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			callCount++
			body, _ := io.ReadAll(req.Body)
			assert.NotContains(t, string(body), "company_id")

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"results":[{"data":[]}],"errors":[]}`)),
				Request:    req,
			}, nil
		}),
	}

	service := handlers.NewNeo4jTopologyService("http://neo4j.local:7474", "neo4j", "secret", "neo4j", client)
	resp, err := service.GetTopology(context.Background(), "all")

	assert.NoError(t, err)
	assert.Empty(t, resp.Hosts)
	assert.GreaterOrEqual(t, callCount, 1)
}
