package transport_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/features/auth/transport"
)

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewConfig(t *testing.T) {
	tests := []struct {
		name     string
		envSetup func(t *testing.T)
		wantCfg  transport.Config
		errMsg   string
	}{
		{
			name: "valid env returns config with parsed values",
			envSetup: func(t *testing.T) {
				t.Helper()
				t.Setenv("AUTH_COOKIE_SECURE", "true")
				t.Setenv("AUTH_COOKIE_DOMAIN", "example.com")
			},
			wantCfg: transport.Config{
				CookieSecure: true,
				CookieDomain: "example.com",
			},
			errMsg: "",
		},
		{
			name: "defaults applied when env unset",
			envSetup: func(t *testing.T) {
				t.Helper()
				unsetEnvForTest(t, "AUTH_COOKIE_SECURE")
				unsetEnvForTest(t, "AUTH_COOKIE_DOMAIN")
			},
			wantCfg: transport.Config{
				CookieSecure: false,
				CookieDomain: "",
			},
			errMsg: "",
		},
		{
			name: "invalid bool value returns error",
			envSetup: func(t *testing.T) {
				t.Helper()
				t.Setenv("AUTH_COOKIE_SECURE", "notabool")
				unsetEnvForTest(t, "AUTH_COOKIE_DOMAIN")
			},
			wantCfg: transport.Config{},
			errMsg:  "process cookie config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			must := require.New(t)

			tt.envSetup(t)

			got, err := transport.NewConfig()

			if tt.errMsg != "" {
				must.Error(err)
				must.ErrorContains(err, tt.errMsg)

				return
			}

			must.NoError(err)
			is.Equal(tt.wantCfg.CookieSecure, got.CookieSecure)
			is.Equal(tt.wantCfg.CookieDomain, got.CookieDomain)
		})
	}
}

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewConfigMust(t *testing.T) {
	t.Run("valid env does not panic and returns config", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		t.Setenv("AUTH_COOKIE_SECURE", "true")
		t.Setenv("AUTH_COOKIE_DOMAIN", "example.com")

		var (
			panicVal any
			config   transport.Config
		)

		func() {
			defer func() {
				panicVal = recover()
			}()

			config = transport.NewConfigMust()
		}()

		must.Nil(panicVal)
		is.True(config.CookieSecure)
		is.Equal("example.com", config.CookieDomain)
	})

	t.Run("invalid env panics with config error", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		t.Setenv("AUTH_COOKIE_SECURE", "notabool")
		unsetEnvForTest(t, "AUTH_COOKIE_DOMAIN")

		var panicVal any

		func() {
			defer func() {
				panicVal = recover()
			}()

			transport.NewConfigMust()
		}()

		must.NotNil(panicVal)

		panicErr, ok := panicVal.(error)
		must.True(ok)

		is.Contains(panicErr.Error(), "get cookie config")
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
