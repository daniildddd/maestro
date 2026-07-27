package middleware

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AllowedOrigins []string `envconfig:"ALLOWED_ORIGINS" required:"true"`
}

func NewConfigMust() Config {
	var cfg Config

	if err := envconfig.Process("MIDDLEWARE", &cfg); err != nil {
		panic(fmt.Errorf("process middleware config: %v", err))
	}

	return cfg
}
