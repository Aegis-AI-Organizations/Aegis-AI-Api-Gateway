package db

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneratePresignedPutURLUsesPublicStorageEndpoint(t *testing.T) {
	t.Setenv("MINIO_ENDPOINT", "aegis-minio-mvp:9000")
	t.Setenv("MINIO_ACCESS_KEY", "minioadmin")
	t.Setenv("MINIO_SECRET_KEY", "minioadmin")
	t.Setenv("MINIO_BUCKET", "aegis-telemetry")
	t.Setenv("MINIO_USE_SSL", "false")

	client, err := NewMinioClient()
	require.NoError(t, err)

	rawURL, err := client.GeneratePresignedPutURL(context.Background(), "agents/a1/payload.json")
	require.NoError(t, err)

	parsedURL, err := url.Parse(rawURL)
	require.NoError(t, err)
	require.Equal(t, "https", parsedURL.Scheme)
	require.Equal(t, PublicMinioEndpoint, parsedURL.Host)
	require.True(t, strings.HasPrefix(rawURL, "https://storage.aegis-ai.fr/"))
}
