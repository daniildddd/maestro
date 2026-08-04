package access

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Secret    string        `envconfig:"SECRET" required:"true"`
	AccessTTL time.Duration `default:"15m"      envconfig:"ACCESS_TTL"`
	Issuer    string        `default:"maestro"  envconfig:"ISSUER"`
}

func NewConfigMust() Config {
	var cfg Config

	if err := envconfig.Process("JWT", &cfg); err != nil {
		panic(fmt.Errorf("process jwt config: %v", err))
	}

	return cfg
}
