package metrics_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/metrics"
)

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewConfigMust(t *testing.T) {
	t.Run("defaults applied when env unset", func(t *testing.T) {
		unsetAllMetricsEnv(t)

		cfg := metrics.NewConfigMust()

		assert.NotEmpty(t, cfg.Buckets)
		assert.Equal(t, ":9100", cfg.Addr)
	})

	t.Run("valid env returns config with parsed values", func(t *testing.T) {
		unsetAllMetricsEnv(t)
		t.Setenv("METRICS_BUCKETS", "0.1,0.5,1")

		cfg := metrics.NewConfigMust()

		assert.Equal(t, []float64{0.1, 0.5, 1}, cfg.Buckets)
	})

	t.Run("invalid BUCKETS panics", func(t *testing.T) {
		unsetAllMetricsEnv(t)
		t.Setenv("METRICS_BUCKETS", "not-a-float")

		mustPanic(t, func() { metrics.NewConfigMust() })
	})
}

func mustPanic(t *testing.T, fn func()) {
	t.Helper()

	var panicVal any

	func() {
		defer func() { panicVal = recover() }()

		fn()
	}()

	must := require.New(t)
	must.NotNil(panicVal)
	err, ok := panicVal.(error)
	must.True(ok)
	must.Error(err)
}

func unsetAllMetricsEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"METRICS_BUCKETS",
		"METRICS_ADDR",
		"METRICS_READ_HEADER_TIMEOUT",
		"METRICS_SHUTDOWN_TIMEOUT",
	}

	for _, key := range keys {
		old, ok := os.LookupEnv(key)
		_ = os.Unsetenv(key) //nolint:errcheck // env var removal cannot fail in practice

		t.Cleanup(func() {
			if ok {
				//nolint:errcheck,usetesting // cannot use t.Setenv inside t.Cleanup; restore original
				_ = os.Setenv(key, old)
			}
		})
	}
}
