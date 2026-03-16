package grpc

import (
	"context"

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
	resp, err := c.PingService.Ping(ctx, &v1.PingRequest{})
	if err != nil {
		return "", err
	}
	return resp.Message, nil
}
