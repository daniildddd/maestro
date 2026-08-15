package pgxadapter_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/repository/postgres/pgxadapter"
)

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewConfig(t *testing.T) {
	t.Run("valid env returns config with parsed values and DSN", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		unsetAllPgxadapterEnv(t)
		t.Setenv("POSTGRES_HOST", "127.0.0.1")
		t.Setenv("POSTGRES_PORT", "5433")
		t.Setenv("POSTGRES_USER", "postgres")
		t.Setenv("POSTGRES_PASSWORD", "pass")
		t.Setenv("POSTGRES_DB", "postgres")
		t.Setenv("POSTGRES_TIMEOUT", "10s")

		cfg, err := pgxadapter.NewConfig()

		must.NoError(err)
		is.Equal("127.0.0.1", cfg.Host)
		is.Equal("5433", cfg.Port)
		is.Equal("postgres", cfg.User)
		is.Equal("pass", cfg.Password)
		is.Equal("postgres", cfg.Database)
		is.Equal(10*time.Second, cfg.Timeout)
		is.Equal("postgres://postgres:pass@127.0.0.1:5433/postgres?sslmode=disable", cfg.DSN)
	})

	t.Run("defaults applied when optional env unset", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		unsetAllPgxadapterEnv(t)
		t.Setenv("POSTGRES_USER", "postgres")
		t.Setenv("POSTGRES_PASSWORD", "pass")

		cfg, err := pgxadapter.NewConfig()

		must.NoError(err)
		is.Equal("localhost", cfg.Host)
		is.Equal("5432", cfg.Port)
		is.Equal("postgres", cfg.Database)
		is.Equal(5*time.Second, cfg.Timeout)
		is.Equal("postgres://postgres:pass@localhost:5432/postgres?sslmode=disable", cfg.DSN)
	})

	t.Run("missing required USER returns error", func(t *testing.T) {
		must := require.New(t)

		unsetAllPgxadapterEnv(t)
		t.Setenv("POSTGRES_PASSWORD", "pass")

		_, err := pgxadapter.NewConfig()

		must.Error(err)
		must.ErrorContains(err, "required key USER missing value")
	})

	t.Run("missing required PASSWORD returns error", func(t *testing.T) {
		must := require.New(t)

		unsetAllPgxadapterEnv(t)
		t.Setenv("POSTGRES_USER", "postgres")

		_, err := pgxadapter.NewConfig()

		must.Error(err)
		must.ErrorContains(err, "required key PASSWORD missing value")
	})
}

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewConfigMust(t *testing.T) {
	t.Run("valid env does not panic and returns config", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		unsetAllPgxadapterEnv(t)
		t.Setenv("POSTGRES_HOST", "127.0.0.1")
		t.Setenv("POSTGRES_PORT", "5433")
		t.Setenv("POSTGRES_USER", "postgres")
		t.Setenv("POSTGRES_PASSWORD", "pass")
		t.Setenv("POSTGRES_DB", "postgres")
		t.Setenv("POSTGRES_TIMEOUT", "10s")

		var cfg pgxadapter.Config

		var panicVal any

		func() {
			defer func() { panicVal = recover() }()

			cfg = pgxadapter.NewConfigMust()
		}()

		must.Nil(panicVal)
		is.Equal("postgres", cfg.User)
		is.Equal("pass", cfg.Password)
		is.NotEmpty(cfg.DSN)
	})

	t.Run("missing required USER panics", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		unsetAllPgxadapterEnv(t)
		t.Setenv("POSTGRES_PASSWORD", "pass")

		var panicVal any

		func() {
			defer func() { panicVal = recover() }()

			pgxadapter.NewConfigMust()
		}()

		must.NotNil(panicVal)
		panicErr, ok := panicVal.(error)
		must.True(ok)
		is.Contains(panicErr.Error(), "get database config")
	})

	t.Run("missing required PASSWORD panics", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		unsetAllPgxadapterEnv(t)
		t.Setenv("POSTGRES_USER", "postgres")

		var panicVal any

		func() {
			defer func() { panicVal = recover() }()

			pgxadapter.NewConfigMust()
		}()

		must.NotNil(panicVal)
		panicErr, ok := panicVal.(error)
		must.True(ok)
		is.Contains(panicErr.Error(), "get database config")
	})
}

func unsetAllPgxadapterEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER", "POSTGRES_PASSWORD",
		"POSTGRES_DB", "POSTGRES_TIMEOUT",
		"USER",
	}
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
