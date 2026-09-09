package cleanup_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/features/auth/cleanup"
)

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewConfigMust(t *testing.T) {
	t.Run("valid env does not panic and returns config", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		unsetAllCleanupEnv(t)
		t.Setenv("AUTH_CLEANUP_INTERVAL", "45m")

		var (
			panicVal any
			config   cleanup.Config
		)

		func() {
			defer func() {
				panicVal = recover()
			}()

			config = cleanup.NewConfigMust()
		}()

		must.Nil(panicVal)
		is.Equal(45*time.Minute, config.Interval)
	})

	t.Run("defaults applied when INTERVAL unset", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		unsetAllCleanupEnv(t)

		var (
			panicVal any
			config   cleanup.Config
		)

		func() {
			defer func() {
				panicVal = recover()
			}()

			config = cleanup.NewConfigMust()
		}()

		must.Nil(panicVal)
		is.Equal(time.Hour, config.Interval)
	})

	t.Run("invalid env panics with config error", func(t *testing.T) {
		unsetAllCleanupEnv(t)
		t.Setenv("AUTH_CLEANUP_INTERVAL", "notaduration")

		mustPanic(t, func() { cleanup.NewConfigMust() })
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

func unsetAllCleanupEnv(t *testing.T) {
	t.Helper()

	keys := []string{"AUTH_CLEANUP_INTERVAL"}
	for _, key := range keys {
		old, ok := os.LookupEnv(key)
		_ = os.Unsetenv(key) //nolint:errcheck // env var removal cannot fail in practice

		t.Cleanup(func() {
			if ok {
				_ = os.Setenv(key, old) //nolint:errcheck,usetesting // cannot use t.Setenv inside t.Cleanup; restore original
			}
		})
	}
}
