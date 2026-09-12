package health

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ReadyTimeout time.Duration `default:"2s" envconfig:"READY_TIMEOUT"`
}

func NewConfigMust() Config {
	var cfg Config

	if err := envconfig.Process("HEALTH", &cfg); err != nil {
		panic(fmt.Errorf("process health config: %w", err))
	}

	return cfg
}
