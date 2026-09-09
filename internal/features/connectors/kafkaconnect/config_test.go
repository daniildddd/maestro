package kafkaconnect_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/features/connectors/kafkaconnect"
)

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewConfigMust(t *testing.T) {
	t.Run("valid env returns config with parsed values", func(t *testing.T) {
		is := assert.New(t)

		unsetAllKafkaConnectEnv(t)
		t.Setenv("KAFKA_CONNECT_BASE_URL", "http://localhost:8083")
		t.Setenv("KAFKA_CONNECT_TIMEOUT", "5s")
		t.Setenv("KAFKA_CONNECT_RETRY_MAX_ATTEMPTS", "3")
		t.Setenv("KAFKA_CONNECT_RETRY_INITIAL_DELAY", "100ms")

		cfg := kafkaconnect.NewConfigMust()

		is.Equal("http://localhost:8083", cfg.BaseURL)
		is.Equal(5*time.Second, cfg.Timeout)
		is.Equal(3, cfg.RetryMaxAttempts)
		is.Equal(100*time.Millisecond, cfg.RetryInitialDelay)
	})

	t.Run("defaults applied when optional env unset", func(t *testing.T) {
		is := assert.New(t)

		unsetAllKafkaConnectEnv(t)
		t.Setenv("KAFKA_CONNECT_BASE_URL", "http://localhost:8083")

		cfg := kafkaconnect.NewConfigMust()

		is.Equal(30*time.Second, cfg.Timeout)
		is.Equal(5, cfg.RetryMaxAttempts)
		is.Equal(300*time.Millisecond, cfg.RetryInitialDelay)
	})

	t.Run("missing required BASE_URL panics", func(t *testing.T) {
		must := require.New(t)

		unsetAllKafkaConnectEnv(t)

		var panicVal any

		func() {
			defer func() { panicVal = recover() }()

			kafkaconnect.NewConfigMust()
		}()

		must.NotNil(panicVal)
		err, ok := panicVal.(error)
		must.True(ok)
		must.Error(err)
	})

	t.Run("invalid TIMEOUT panics", func(t *testing.T) {
		must := require.New(t)

		unsetAllKafkaConnectEnv(t)
		t.Setenv("KAFKA_CONNECT_BASE_URL", "http://localhost:8083")
		t.Setenv("KAFKA_CONNECT_TIMEOUT", "not-a-duration")

		var panicVal any

		func() {
			defer func() { panicVal = recover() }()

			kafkaconnect.NewConfigMust()
		}()

		must.NotNil(panicVal)
		err, ok := panicVal.(error)
		must.True(ok)
		must.Error(err)
	})
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
				_ = os.Setenv(key, old) //nolint:errcheck,usetesting // cannot use t.Setenv inside t.Cleanup; restore original
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

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
				RetryInitialDelay: 5 * timeMillisecond,
			},
		},
		{
			name: "non positive timeout rejected",
			config: kafkaconnect.Config{
				BaseURL:           "http://localhost:8083",
				Timeout:           0,
				RetryMaxAttempts:  3,
				RetryInitialDelay: 5 * timeMillisecond,
			},
			wantErr: true,
		},
		{
			name: "non positive retry attempts rejected",
			config: kafkaconnect.Config{
				BaseURL:           "http://localhost:8083",
				Timeout:           defaultTestTimeout,
				RetryMaxAttempts:  0,
				RetryInitialDelay: 5 * timeMillisecond,
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			err := tt.config.Validate()

			if tt.wantErr {
				must.Error(err)

				return
			}

			must.NoError(err)
		})
	}
}
