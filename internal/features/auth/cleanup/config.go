package cleanup

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Interval time.Duration `default:"1h" envconfig:"INTERVAL"`
}

func NewConfigMust() Config {
	var cfg Config

	if err := envconfig.Process("AUTH_CLEANUP", &cfg); err != nil {
		panic(fmt.Errorf("get cleanup config: %w", err))
	}

	return cfg
}
