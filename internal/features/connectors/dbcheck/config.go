package dbcheck

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	StepTimeout    time.Duration `default:"10s" envconfig:"STEP_TIMEOUT"`
	MaxTableChecks int           `default:"50"  envconfig:"MAX_TABLE_CHECKS"`
}

func NewConfigMust() Config {
	var cfg Config

	if err := envconfig.Process("VALIDATE_DB", &cfg); err != nil {
		panic(fmt.Errorf("process validate db config: %w", err))
	}

	return cfg
}
