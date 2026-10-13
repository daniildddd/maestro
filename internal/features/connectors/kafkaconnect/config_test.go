package kafkaconnect_test

import (
	"os"
	"testing"
	"time"

	gotest "github.com/allure-framework/allure-go/commons/gotest" //nolint:revive // package clause is allure, gotest keeps call sites readable
	allureassert "github.com/allure-framework/allure-go/testify/assert"
	allurerequire "github.com/allure-framework/allure-go/testify/require"

	"github.com/daniildddd/maestro/internal/features/connectors/kafkaconnect"
)

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewConfigMust(t *testing.T) {
	gotest.Test(t, "new config must", func(a *gotest.Context) {
		a.Step("valid env returns config with parsed values", func(a *gotest.Context) {
			is := allureassert.New(a)

			unsetAllKafkaConnectEnv(a.T())
			a.T().Setenv("KAFKA_CONNECT_BASE_URL", "http://localhost:8083")
			a.T().Setenv("KAFKA_CONNECT_TIMEOUT", "5s")
			a.T().Setenv("KAFKA_CONNECT_RETRY_MAX_ATTEMPTS", "3")
			a.T().Setenv("KAFKA_CONNECT_RETRY_INITIAL_DELAY", "100ms")

			cfg := kafkaconnect.NewConfigMust()

			is.Equal("http://localhost:8083", cfg.BaseURL)
			is.Equal(5*time.Second, cfg.Timeout)
			is.Equal(3, cfg.RetryMaxAttempts)
			is.Equal(100*time.Millisecond, cfg.RetryInitialDelay)
		})

		a.Step("defaults applied when optional env unset", func(a *gotest.Context) {
			is := allureassert.New(a)

			unsetAllKafkaConnectEnv(a.T())
			a.T().Setenv("KAFKA_CONNECT_BASE_URL", "http://localhost:8083")

			cfg := kafkaconnect.NewConfigMust()

			is.Equal(30*time.Second, cfg.Timeout)
			is.Equal(5, cfg.RetryMaxAttempts)
			is.Equal(300*time.Millisecond, cfg.RetryInitialDelay)
		})

		a.Step("missing required BASE_URL panics", func(a *gotest.Context) {
			unsetAllKafkaConnectEnv(a.T())

			mustPanic(a.T(), func() { kafkaconnect.NewConfigMust() })
		})

		a.Step("invalid TIMEOUT panics", func(a *gotest.Context) {
			unsetAllKafkaConnectEnv(a.T())
			a.T().Setenv("KAFKA_CONNECT_BASE_URL", "http://localhost:8083")
			a.T().Setenv("KAFKA_CONNECT_TIMEOUT", "not-a-duration")

			mustPanic(a.T(), func() { kafkaconnect.NewConfigMust() })
		})

		a.Step("zero TIMEOUT panics on validation", func(a *gotest.Context) {
			unsetAllKafkaConnectEnv(a.T())
			a.T().Setenv("KAFKA_CONNECT_BASE_URL", "http://localhost:8083")
			a.T().Setenv("KAFKA_CONNECT_TIMEOUT", "0s")

			mustPanic(a.T(), func() { kafkaconnect.NewConfigMust() })
		})

		a.Step("zero RETRY_MAX_ATTEMPTS panics on validation", func(a *gotest.Context) {
			unsetAllKafkaConnectEnv(a.T())
			a.T().Setenv("KAFKA_CONNECT_BASE_URL", "http://localhost:8083")
			a.T().Setenv("KAFKA_CONNECT_RETRY_MAX_ATTEMPTS", "0")

			mustPanic(a.T(), func() { kafkaconnect.NewConfigMust() })
		})
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Config"),
		gotest.WithSeverity("minor"),
		gotest.WithOwner("qa-intern"),
	)
}

func mustPanic(t *testing.T, fn func()) {
	t.Helper()

	var panicVal any

	func() {
		defer func() { panicVal = recover() }()

		fn()
	}()

	must := allurerequire.New(t)
	must.NotNil(panicVal)
	err, ok := panicVal.(error)
	must.True(ok)
	must.Error(err)
}

func unsetAllKafkaConnectEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"KAFKA_CONNECT_BASE_URL", "KAFKA_CONNECT_TIMEOUT",
		"KAFKA_CONNECT_RETRY_MAX_ATTEMPTS", "KAFKA_CONNECT_RETRY_INITIAL_DELAY",
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

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "config validate", func(a *gotest.Context) {
		tests := []struct {
			name    string
			config  kafkaconnect.Config
			wantErr bool
		}{
			{
				name: "valid config passes",
				config: kafkaconnect.Config{
					BaseURL:           "http://localhost:8083",
					Timeout:           defaultTestTimeout,
					RetryMaxAttempts:  3,
					RetryInitialDelay: 5 * retryDelayUnit,
				},
			},
			{
				name: "non positive timeout rejected",
				config: kafkaconnect.Config{
					BaseURL:           "http://localhost:8083",
					Timeout:           0,
					RetryMaxAttempts:  3,
					RetryInitialDelay: 5 * retryDelayUnit,
				},
				wantErr: true,
			},
			{
				name: "non positive retry attempts rejected",
				config: kafkaconnect.Config{
					BaseURL:           "http://localhost:8083",
					Timeout:           defaultTestTimeout,
					RetryMaxAttempts:  0,
					RetryInitialDelay: 5 * retryDelayUnit,
				},
				wantErr: true,
			},
			{
				name: "non positive retry delay rejected",
				config: kafkaconnect.Config{
					BaseURL:           "http://localhost:8083",
					Timeout:           defaultTestTimeout,
					RetryMaxAttempts:  3,
					RetryInitialDelay: 0,
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

				err := tt.config.Validate()

				if tt.wantErr {
					must.Error(err)

					return
				}

				must.NoError(err)
			})
		}
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Config"),
		gotest.WithSeverity("minor"),
		gotest.WithOwner("qa-intern"),
	)
}
