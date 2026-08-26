package kafkaconnect

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	BaseURL string        `envconfig:"BASE_URL" required:"true"`
	Timeout time.Duration `default:"30s"        envconfig:"TIMEOUT"`
}

func NewConfigMust() Config {
	var cfg Config

	if err := envconfig.Process("KAFKA_CONNECT", &cfg); err != nil {
		panic(fmt.Errorf("process kafka connect config: %w", err))
	}

	return cfg
}
