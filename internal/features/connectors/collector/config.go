package collector

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type CollectorConfig struct {
	Interval time.Duration `default:"15s" envconfig:"INTERVAL"`
	Timeout  time.Duration `default:"10s" envconfig:"TIMEOUT"`
}

func NewCollectorConfigMust() CollectorConfig {
	var cfg CollectorConfig

	if err := envconfig.Process("METRICS_COLLECTOR", &cfg); err != nil {
		panic(fmt.Errorf("process metrics collector config: %w", err))
	}

	return cfg
}
