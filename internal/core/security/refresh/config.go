package refresh

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	TTL time.Duration `default:"720h" envconfig:"TTL"`
}

func NewConfigMust() Config {
	var cfg Config

	if err := envconfig.Process("REFRESH", &cfg); err != nil {
		panic(fmt.Errorf("process refresh token config: %v", err))
	}

	return cfg
}
