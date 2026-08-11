package access_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/security/access"
)

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewConfigMust(t *testing.T) {
	t.Run("valid env returns config with parsed values", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		unsetAllAccessEnv(t)
		t.Setenv("JWT_SECRET", "super-secret-key")
		t.Setenv("JWT_ACCESS_TTL", "30m")
		t.Setenv("JWT_ISSUER", "maestro-test")

		cfg := access.NewConfigMust()

		must.NotEmpty(cfg)
		is.Equal("super-secret-key", cfg.Secret)
		is.Equal(30*time.Minute, cfg.AccessTTL)
		is.Equal("maestro-test", cfg.Issuer)
	})

	t.Run("defaults applied when optional env unset", func(t *testing.T) {
		is := assert.New(t)

		unsetAllAccessEnv(t)
		t.Setenv("JWT_SECRET", "super-secret-key")

		cfg := access.NewConfigMust()

		is.Equal("maestro", cfg.Issuer)
		is.Equal(15*time.Minute, cfg.AccessTTL)
	})

	t.Run("missing required SECRET panics", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		unsetAllAccessEnv(t)

		var panicVal any

		func() {
			defer func() { panicVal = recover() }()

			access.NewConfigMust()
		}()

		must.NotNil(panicVal)
		err, ok := panicVal.(error)
		must.True(ok)
		is.ErrorContains(err, "required key SECRET missing value")
	})
}

func unsetAllAccessEnv(t *testing.T) {
	t.Helper()

	keys := []string{"JWT_SECRET", "JWT_ACCESS_TTL", "JWT_ISSUER"}
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
