package kafkaconnect

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	BaseURL           string        `envconfig:"BASE_URL" required:"true"`
	Timeout           time.Duration `default:"30s"        envconfig:"TIMEOUT"`
	RetryMaxAttempts  int           `default:"5"          envconfig:"RETRY_MAX_ATTEMPTS"`
	RetryInitialDelay time.Duration `default:"300ms"      envconfig:"RETRY_INITIAL_DELAY"`
}

func NewConfigMust() Config {
	var cfg Config

	if err := envconfig.Process("KAFKA_CONNECT", &cfg); err != nil {
		panic(fmt.Errorf("process kafka connect config: %w", err))
	}

	if err := cfg.Validate(); err != nil {
		panic(fmt.Errorf("validate kafka connect config: %w", err))
	}

	return cfg
}

func (c Config) Validate() error {
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %s", c.Timeout)
	}

	if c.RetryMaxAttempts <= 0 {
		return fmt.Errorf("retry max attempts must be positive, got %d", c.RetryMaxAttempts)
	}

	if c.RetryInitialDelay <= 0 {
		return fmt.Errorf("retry initial delay must be positive, got %s", c.RetryInitialDelay)
	}

	return nil
}
