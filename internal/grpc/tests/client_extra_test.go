package grpc_test

import (
	"context"
	"testing"

	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/grpc"
	"github.com/stretchr/testify/assert"
)

func TestClient_Close_NilConn(t *testing.T) {
	client := &agrpc.Client{}
	err := client.Close()
	assert.NoError(t, err)
}

func TestClient_ServiceErrors_Nil(t *testing.T) {
	client := &agrpc.Client{}
	ctx := context.Background()

	_, err := client.Ping(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = client.StartScan(ctx, "img")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = client.GetScanStatus(ctx, "s1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = client.ListScans(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = client.GetScanReport(ctx, "s1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = client.GetVulnerabilities(ctx, "s1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = client.GetEvidences(ctx, "v1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = client.WatchScanStatus(ctx, "s1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}
