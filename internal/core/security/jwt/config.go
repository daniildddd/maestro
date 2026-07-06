package jwt

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Secret    string        `envconfig:"SECRET" required:"true"`
	AccessTTL time.Duration `envconfig:"ACCESS_TTL" default:"15m"`
	Issuer    string        `envconfig:"ISSUER" default:"maestro"`
}

func NewConfigMust() Config {
	var cfg Config
	if err := envconfig.Process("JWT", &cfg); err != nil {
		panic(fmt.Errorf("process jwt config: %v", err))
	}

	return cfg
}
