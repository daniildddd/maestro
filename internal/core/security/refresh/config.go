package refresh

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	TTL time.Duration `envconfig:"TTL" default:"720h"`
}

func NewConfigMust() Config {
	var cfg Config
	if err := envconfig.Process("REFRESH", &cfg); err != nil {
		panic(fmt.Errorf("process refresh token config: %v", err))
	}

	return cfg
}
