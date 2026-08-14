package middleware_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/transport/middleware"
)

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewConfigMust(t *testing.T) {
	t.Run("valid env returns config with parsed origins and roles", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		unsetAllMiddlewareEnv(t)
		t.Setenv("MIDDLEWARE_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")
		t.Setenv("MIDDLEWARE_ROLES", "admin,user")

		cfg := middleware.NewConfigMust()

		must.NotEmpty(cfg)
		is.Equal([]string{"http://localhost:3000", "http://localhost:5173"}, cfg.AllowedOrigins)
		is.Equal([]string{"admin", "user"}, cfg.Roles)
	})

	t.Run("single origin and role parsed correctly", func(t *testing.T) {
		is := assert.New(t)

		unsetAllMiddlewareEnv(t)
		t.Setenv("MIDDLEWARE_ALLOWED_ORIGINS", "http://localhost:3000")
		t.Setenv("MIDDLEWARE_ROLES", "admin")

		cfg := middleware.NewConfigMust()

		is.Equal([]string{"http://localhost:3000"}, cfg.AllowedOrigins)
		is.Equal([]string{"admin"}, cfg.Roles)
	})

	t.Run("missing required ALLOWED_ORIGINS panics", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		unsetAllMiddlewareEnv(t)
		t.Setenv("MIDDLEWARE_ROLES", "admin")

		var panicVal any

		func() {
			defer func() { panicVal = recover() }()

			middleware.NewConfigMust()
		}()

		must.NotNil(panicVal)
		err, ok := panicVal.(error)
		must.True(ok)
		is.ErrorContains(err, "required key ALLOWED_ORIGINS missing value")
	})

	t.Run("missing required ROLES panics", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		unsetAllMiddlewareEnv(t)
		t.Setenv("MIDDLEWARE_ALLOWED_ORIGINS", "http://localhost:3000")

		var panicVal any

		func() {
			defer func() { panicVal = recover() }()

			middleware.NewConfigMust()
		}()

		must.NotNil(panicVal)
		err, ok := panicVal.(error)
		must.True(ok)
		is.ErrorContains(err, "required key ROLES missing value")
	})
}

func unsetAllMiddlewareEnv(t *testing.T) {
	t.Helper()

	keys := []string{"MIDDLEWARE_ALLOWED_ORIGINS", "MIDDLEWARE_ROLES"}
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
