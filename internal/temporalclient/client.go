package temporalclient

import (
	"fmt"
	"os"

	"go.temporal.io/sdk/client"
)

func Connect() (client.Client, error) {
	host := os.Getenv("TEMPORAL_HOST")
	namespace := os.Getenv("TEMPORAL_NAMESPACE")

	if host == "" {
		host = "localhost:7233"
	}
	if namespace == "" {
		namespace = "default"
	}

	c, err := client.Dial(client.Options{
		HostPort:  host,
		Namespace: namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal client: %w", err)
	}

	return c, nil
}
