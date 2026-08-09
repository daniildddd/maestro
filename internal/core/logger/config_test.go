package logger_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/logger"
)

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewConfig(t *testing.T) {
	tests := []struct {
		name     string
		envSetup func(t *testing.T)
		wantCfg  logger.Config
		errMsg   string
	}{
		{
			name: "valid env returns config",
			envSetup: func(t *testing.T) {
				t.Helper()
				t.Setenv("LOGGER_LEVEL", "DEBUG")
				t.Setenv("LOGGER_FOLDER", "/test")
			},
			wantCfg: logger.Config{
				Level:  "DEBUG",
				Folder: "/test",
			},
			errMsg: "",
		},
		{
			name: "missing required FOLDER returns error",
			envSetup: func(t *testing.T) {
				t.Helper()
				t.Setenv("LOGGER_LEVEL", "DEBUG")
				unsetEnvForTest(t, "LOGGER_FOLDER")
			},
			wantCfg: logger.Config{},
			errMsg:  "FOLDER missing",
		},
		{
			name: "default level applied when LOGGER_LEVEL unset",
			envSetup: func(t *testing.T) {
				t.Helper()
				t.Setenv("LOGGER_FOLDER", "/test")
				unsetEnvForTest(t, "LOGGER_LEVEL")
			},
			wantCfg: logger.Config{
				Level:  "DEBUG",
				Folder: "/test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			must := require.New(t)

			tt.envSetup(t)

			got, err := logger.NewConfig()

			if tt.errMsg != "" {
				must.Error(err)
				must.ErrorContains(err, tt.errMsg)

				return
			}

			must.NoError(err)
			is.Equal(tt.wantCfg.Folder, got.Folder)
			is.Equal(tt.wantCfg.Level, got.Level)
		})
	}
}

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewConfigMust(t *testing.T) {
	t.Run("valid env does not panic and returns config", func(t *testing.T) {
		must := require.New(t)
		is := assert.New(t)
		t.Setenv("LOGGER_LEVEL", "DEBUG")
		t.Setenv("LOGGER_FOLDER", "/test")

		var (
			panicVal any
			config   logger.Config
		)

		func() {
			defer func() {
				panicVal = recover()
			}()

			config = logger.NewConfigMust()
		}()

		must.Nil(panicVal)
		is.Equal("DEBUG", config.Level)
		is.Equal("/test", config.Folder)
	})
	t.Run("missing required FOLDER panics with FOLDER error", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		t.Setenv("LOGGER_LEVEL", "DEBUG")
		unsetEnvForTest(t, "LOGGER_FOLDER")

		var panicVal any

		func() {
			defer func() {
				panicVal = recover()
			}()

			logger.NewConfigMust()
		}()

		must.NotNil(panicVal)

		err, ok := panicVal.(error)
		must.True(ok)

		is.ErrorContains(err, "FOLDER missing")
	})
}

func unsetEnvForTest(t *testing.T, key string) {
	t.Helper()

	old, ok := os.LookupEnv(key)
	_ = os.Unsetenv(key) //nolint:errcheck // env var removal cannot fail in practice

	t.Cleanup(func() {
		if ok {
			_ = os.Setenv(key, old) //nolint:errcheck,usetesting // cannot use t.Setenv inside t.Cleanup (test already finished); restore original
		}
	})
}
