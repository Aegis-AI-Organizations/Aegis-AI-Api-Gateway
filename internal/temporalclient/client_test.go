package temporalclient

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnect_DialError(t *testing.T) {
	_ = os.Setenv("TEMPORAL_HOST", "invalid-host:1234")

	c, err := Connect()
	assert.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "unable to create temporal client")
}

func TestConnect_Defaults(t *testing.T) {
	_ = os.Unsetenv("TEMPORAL_HOST")
	_ = os.Unsetenv("TEMPORAL_NAMESPACE")
}
