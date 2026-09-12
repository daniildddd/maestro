package health_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/transport/health"
)

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewConfigMust(t *testing.T) {
	t.Run("valid env does not panic and returns config", func(t *testing.T) {
		must := require.New(t)

		t.Setenv("HEALTH_READY_TIMEOUT", "5s")

		var (
			panicVal any
			config   health.Config
		)

		func() {
			defer func() {
				panicVal = recover()
			}()

			config = health.NewConfigMust()
		}()

		must.Nil(panicVal)
		must.Equal(5*time.Second, config.ReadyTimeout)
	})

	t.Run("defaults applied when env unset", func(t *testing.T) {
		must := require.New(t)

		unsetEnvForTest(t, "HEALTH_READY_TIMEOUT")

		var (
			panicVal any
			config   health.Config
		)

		func() {
			defer func() {
				panicVal = recover()
			}()

			config = health.NewConfigMust()
		}()

		must.Nil(panicVal)
		must.Equal(2*time.Second, config.ReadyTimeout)
	})

	t.Run("invalid env panics with config error", func(t *testing.T) {
		t.Setenv("HEALTH_READY_TIMEOUT", "notaduration")

		mustPanic(t, func() { health.NewConfigMust() })
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

func unsetEnvForTest(t *testing.T, key string) {
	t.Helper()

	old, ok := os.LookupEnv(key)
	_ = os.Unsetenv(key) //nolint:errcheck // env var removal cannot fail in practice

	t.Cleanup(func() {
		if ok {
			//nolint:errcheck,usetesting // cannot use t.Setenv inside t.Cleanup
			_ = os.Setenv(key, old)
		}
	})
}
