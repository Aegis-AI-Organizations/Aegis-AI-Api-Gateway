package grpc

import (
	"context"
	"fmt"
	"testing"

	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/grpc/aegis/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type MockPingServiceClient struct {
	mock.Mock
}

func (m *MockPingServiceClient) Ping(ctx context.Context, in *v1.PingRequest, opts ...grpc.CallOption) (*v1.PingResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*v1.PingResponse), args.Error(1)
}

type MockScanServiceClient struct {
	mock.Mock
}

func (m *MockScanServiceClient) StartScan(ctx context.Context, in *v1.StartScanRequest, opts ...grpc.CallOption) (*v1.StartScanResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*v1.StartScanResponse), args.Error(1)
}

func (m *MockScanServiceClient) GetScanStatus(ctx context.Context, in *v1.GetScanStatusRequest, opts ...grpc.CallOption) (*v1.GetScanStatusResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*v1.GetScanStatusResponse), args.Error(1)
}

type MockVulnerabilityServiceClient struct {
	mock.Mock
}

func (m *MockVulnerabilityServiceClient) GetVulnerabilities(ctx context.Context, in *v1.GetVulnerabilitiesRequest, opts ...grpc.CallOption) (*v1.GetVulnerabilitiesResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*v1.GetVulnerabilitiesResponse), args.Error(1)
}

func TestClient_Ping(t *testing.T) {
	mockPing := new(MockPingServiceClient)
	client := &Client{
		PingService: mockPing,
	}

	mockPing.On("Ping", mock.Anything, mock.Anything).Return(&v1.PingResponse{Message: "pong"}, nil)

	msg, err := client.Ping(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "pong", msg)
}

func TestClient_ScanServices(t *testing.T) {
	mockScan := new(MockScanServiceClient)
	mockVuln := new(MockVulnerabilityServiceClient)
	client := &Client{
		ScanService:          mockScan,
		VulnerabilityService: mockVuln,
	}

	mockScan.On("StartScan", mock.Anything, mock.Anything).Return(&v1.StartScanResponse{ScanId: "s1"}, nil)
	mockScan.On("GetScanStatus", mock.Anything, mock.Anything).Return(&v1.GetScanStatusResponse{Status: "RUNNING"}, nil)
	mockVuln.On("GetVulnerabilities", mock.Anything, mock.Anything).Return(&v1.GetVulnerabilitiesResponse{Vulnerabilities: []*v1.Vulnerability{}}, nil)

	id, err := client.StartScan(context.Background(), "img")
	assert.NoError(t, err)
	assert.Equal(t, "s1", id)

	status, err := client.GetScanStatus(context.Background(), "s1")
	assert.NoError(t, err)
	assert.Equal(t, "RUNNING", status)

	vulns, err := client.GetVulnerabilities(context.Background(), "s1")
	assert.NoError(t, err)
	assert.Len(t, vulns, 0)
}

func TestClient_Failures(t *testing.T) {
	mockScan := new(MockScanServiceClient)
	mockVuln := new(MockVulnerabilityServiceClient)
	client := &Client{
		ScanService:          mockScan,
		VulnerabilityService: mockVuln,
	}

	mockScan.On("StartScan", mock.Anything, mock.Anything).Return((*v1.StartScanResponse)(nil), fmt.Errorf("rpc error"))
	mockScan.On("GetScanStatus", mock.Anything, mock.Anything).Return((*v1.GetScanStatusResponse)(nil), fmt.Errorf("rpc error"))
	mockVuln.On("GetVulnerabilities", mock.Anything, mock.Anything).Return((*v1.GetVulnerabilitiesResponse)(nil), fmt.Errorf("rpc error"))

	_, err := client.StartScan(context.Background(), "img")
	assert.Error(t, err)

	_, err = client.GetScanStatus(context.Background(), "s1")
	assert.Error(t, err)

	_, err = client.GetVulnerabilities(context.Background(), "s1")
	assert.Error(t, err)
}

func TestClient_PingNil(t *testing.T) {
    client := &Client{}
    _, err := client.Ping(context.Background())
    assert.Error(t, err)
}

func TestNewClient(t *testing.T) {
	// Should succeed in creating the structure even if connection is lazy/not established yet
	c, err := NewClient("localhost:1234")
	assert.NoError(t, err)
	assert.NotNil(t, c)
	defer func() { _ = c.Close() }()
}
