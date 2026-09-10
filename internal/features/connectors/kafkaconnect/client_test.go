package kafkaconnect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/features/connectors/plugins"
)

func TestNewHTTPClient(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	registry := plugins.NewRegistry(plugins.PostgresAdapter{})

	client := NewHTTPClient(Config{
		BaseURL:           "http://connect:8083",
		Timeout:           time.Second,
		RetryMaxAttempts:  5,
		RetryInitialDelay: 300 * time.Millisecond,
	}, registry)

	must.Equal("http://connect:8083", client.baseURL)
	must.Equal(time.Second, client.client.Timeout)
	must.Same(registry, client.registry)
	must.Equal(5, client.retryMaxAttempts)
	must.Equal(300*time.Millisecond, client.retryInitialDelay)
}
