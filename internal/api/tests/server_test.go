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

func init() {
	// Set JWT_SECRET for all tests in this package to avoid auth middleware warnings/errors
	os.Setenv("JWT_SECRET", "test-secret-key-12345")
}

func TestStart(t *testing.T) {
	// Use a specific port for the test
	port := "18081"
	os.Setenv("PORT", port)
	defer os.Unsetenv("PORT")

	client := &agrpc.Client{} // Nil service clients are fine for NewRouter

	// We run Start in a goroutine as it blocks
	go func() {
		api.Start(client)
	}()

	// Wait for the server to be ready with a timeout
	maxRetries := 10
	var resp *http.Response
	var err error
	for i := 0; i < maxRetries; i++ {
		resp, err = http.Get("http://localhost:" + port + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	assert.NoError(t, err)
	if err == nil {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
}
