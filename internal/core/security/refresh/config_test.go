package refresh_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/security/refresh"
)

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewConfigMust(t *testing.T) {
	t.Run("valid env returns config with parsed TTL", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		unsetAllRefreshEnv(t)
		t.Setenv("REFRESH_TTL", "48h")

		cfg := refresh.NewConfigMust()

		must.NotEmpty(cfg)
		is.Equal(48*time.Hour, cfg.TTL)
	})

	t.Run("defaults applied when REFRESH_TTL unset", func(t *testing.T) {
		is := assert.New(t)

		unsetAllRefreshEnv(t)

		cfg := refresh.NewConfigMust()

		is.Equal(720*time.Hour, cfg.TTL)
	})

	t.Run("invalid TTL panics", func(t *testing.T) {
		unsetAllRefreshEnv(t)
		t.Setenv("REFRESH_TTL", "not-a-duration")

		mustPanic(t, func() { refresh.NewConfigMust() })
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

func unsetAllRefreshEnv(t *testing.T) {
	t.Helper()

	keys := []string{"REFRESH_TTL"}
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
