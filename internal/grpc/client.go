package grpc

import (
	"context"
	"fmt"

	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/grpc/aegis/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn                 *grpc.ClientConn
	PingService          v1.PingServiceClient
	ScanService          v1.ScanServiceClient
	VulnerabilityService v1.VulnerabilityServiceClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:                 conn,
		PingService:          v1.NewPingServiceClient(conn),
		ScanService:          v1.NewScanServiceClient(conn),
		VulnerabilityService: v1.NewVulnerabilityServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	if c.PingService == nil {
		return "", fmt.Errorf("ping service not initialized")
	}
	resp, err := c.PingService.Ping(ctx, &v1.PingRequest{})
	if err != nil {
		return "", err
	}
	return resp.Message, nil
}

func (c *Client) StartScan(ctx context.Context, image string) (string, error) {
	if c.ScanService == nil {
		return "", fmt.Errorf("scan service not initialized")
	}
	resp, err := c.ScanService.StartScan(ctx, &v1.StartScanRequest{TargetImage: image})
	if err != nil {
		return "", err
	}
	return resp.ScanId, nil
}

func (c *Client) GetScanStatus(ctx context.Context, scanID string) (string, error) {
	if c.ScanService == nil {
		return "", fmt.Errorf("scan service not initialized")
	}
	resp, err := c.ScanService.GetScanStatus(ctx, &v1.GetScanStatusRequest{ScanId: scanID})
	if err != nil {
		return "", err
	}
	return resp.Status, nil
}

func (c *Client) ListScans(ctx context.Context) ([]*v1.ScanDetails, error) {
	if c.ScanService == nil {
		return nil, fmt.Errorf("scan service not initialized")
	}
	resp, err := c.ScanService.ListScans(ctx, &v1.ListScansRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Scans, nil
}

func (c *Client) GetScanReport(ctx context.Context, scanID string) ([]byte, error) {
	if c.ScanService == nil {
		return nil, fmt.Errorf("scan service not initialized")
	}
	resp, err := c.ScanService.GetScanReport(ctx, &v1.GetScanReportRequest{ScanId: scanID})
	if err != nil {
		return nil, err
	}
	return resp.PdfData, nil
}

func (c *Client) GetVulnerabilities(ctx context.Context, scanID string) ([]*v1.Vulnerability, error) {
	if c.VulnerabilityService == nil {
		return nil, fmt.Errorf("vulnerability service not initialized")
	}
	resp, err := c.VulnerabilityService.GetVulnerabilities(ctx, &v1.GetVulnerabilitiesRequest{ScanId: scanID})
	if err != nil {
		return nil, err
	}
	return resp.Vulnerabilities, nil
}

func (c *Client) GetEvidences(ctx context.Context, vulnID string) ([]*v1.Evidence, error) {
	if c.VulnerabilityService == nil {
		return nil, fmt.Errorf("vulnerability service not initialized")
	}
	resp, err := c.VulnerabilityService.GetEvidences(ctx, &v1.GetEvidencesRequest{VulnerabilityId: vulnID})
	if err != nil {
		return nil, err
	}
	return resp.Evidences, nil
}
