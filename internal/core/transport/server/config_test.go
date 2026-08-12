package server_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/transport/server"
)

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewConfigMust(t *testing.T) {
	t.Run("valid env returns config with parsed values", func(t *testing.T) {
		is := assert.New(t)

		unsetAllServerEnv(t)
		t.Setenv("SERVER_ADDR", "0.0.0.0:9090")
		t.Setenv("SERVER_READ_TIMEOUT", "5s")
		t.Setenv("SERVER_READ_HEADER_TIMEOUT", "2s")
		t.Setenv("SERVER_WRITE_TIMEOUT", "10s")
		t.Setenv("SERVER_IDLE_TIMEOUT", "60s")
		t.Setenv("SERVER_MAX_HEADER_BYTES", "4096")
		t.Setenv("SERVER_DISABLE_GENERAL_OPTIONS_HANDLER", "false")
		t.Setenv("SERVER_SHUTDOWN_TIMEOUT", "15s")

		cfg := server.NewConfigMust()

		is.Equal("0.0.0.0:9090", cfg.Addr)
		is.Equal(5*time.Second, cfg.ReadTimeout)
		is.Equal(2*time.Second, cfg.ReadHeaderTimeout)
		is.Equal(10*time.Second, cfg.WriteTimeout)
		is.Equal(60*time.Second, cfg.IdleTimeout)
		is.Equal(4096, cfg.MaxHeaderBytes)
		is.False(cfg.DisableGeneralOptionsHandler)
		is.Equal(15*time.Second, cfg.ShutdownTimeout)
	})

	t.Run("defaults applied when env unset", func(t *testing.T) {
		is := assert.New(t)

		unsetAllServerEnv(t)

		cfg := server.NewConfigMust()

		is.Equal("localhost:8080", cfg.Addr)
		is.Equal(30*time.Second, cfg.ReadTimeout)
		is.Equal(10*time.Second, cfg.ReadHeaderTimeout)
		is.Equal(30*time.Second, cfg.WriteTimeout)
		is.Equal(120*time.Second, cfg.IdleTimeout)
		is.Equal(8192, cfg.MaxHeaderBytes)
		is.True(cfg.DisableGeneralOptionsHandler)
		is.Equal(30*time.Second, cfg.ShutdownTimeout)
	})

	t.Run("invalid duration panics", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		unsetAllServerEnv(t)
		t.Setenv("SERVER_READ_TIMEOUT", "not-a-duration")

		var panicVal any

		func() {
			defer func() { panicVal = recover() }()

			server.NewConfigMust()
		}()

		must.NotNil(panicVal)
		err, ok := panicVal.(error)
		must.True(ok)
		is.ErrorContains(err, "process server config")
	})
}

func unsetAllServerEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"SERVER_ADDR", "SERVER_READ_TIMEOUT", "SERVER_READ_HEADER_TIMEOUT",
		"SERVER_WRITE_TIMEOUT", "SERVER_IDLE_TIMEOUT", "SERVER_MAX_HEADER_BYTES",
		"SERVER_DISABLE_GENERAL_OPTIONS_HANDLER", "SERVER_SHUTDOWN_TIMEOUT",
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
