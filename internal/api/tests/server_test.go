package api_test

import (
	"net/http"
	"os"
	"testing"
	"time"

	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api"
	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	// Use a random port to avoid collisions
	os.Setenv("PORT", "18080")
	defer os.Unsetenv("PORT")

	client := &agrpc.Client{} // Nil service clients are fine for NewRouter

	// We run Start in a goroutine as it blocks
	go func() {
		api.Start(client)
	}()

	// Give it a moment to start
	time.Sleep(200 * time.Millisecond)

	// Verify we can reach the health endpoint
	resp, err := http.Get("http://localhost:18080/health")
	assert.NoError(t, err)
	if err == nil {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
}
