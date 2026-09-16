package collector_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/features/connectors/collector"
)

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewCollectorConfigMust(t *testing.T) {
	t.Run("valid env returns config with parsed values", func(t *testing.T) {
		is := assert.New(t)

		unsetAllCollectorEnv(t)
		t.Setenv("METRICS_COLLECTOR_INTERVAL", "30s")
		t.Setenv("METRICS_COLLECTOR_TIMEOUT", "5s")

		cfg := collector.NewCollectorConfigMust()

		is.Equal(30*time.Second, cfg.Interval)
		is.Equal(5*time.Second, cfg.Timeout)
	})

	t.Run("defaults applied when env unset", func(t *testing.T) {
		is := assert.New(t)

		unsetAllCollectorEnv(t)

		cfg := collector.NewCollectorConfigMust()

		is.Equal(15*time.Second, cfg.Interval)
		is.Equal(10*time.Second, cfg.Timeout)
	})

	t.Run("invalid INTERVAL panics", func(t *testing.T) {
		unsetAllCollectorEnv(t)
		t.Setenv("METRICS_COLLECTOR_INTERVAL", "not-a-duration")

		mustPanic(t, func() { collector.NewCollectorConfigMust() })
	})

	t.Run("invalid TIMEOUT panics", func(t *testing.T) {
		unsetAllCollectorEnv(t)
		t.Setenv("METRICS_COLLECTOR_TIMEOUT", "not-a-duration")

		mustPanic(t, func() { collector.NewCollectorConfigMust() })
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

func unsetAllCollectorEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"METRICS_COLLECTOR_INTERVAL", "METRICS_COLLECTOR_TIMEOUT",
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
