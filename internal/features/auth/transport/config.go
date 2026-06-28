package transport

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	CookieSecure bool   `json:"COOKIE_SECURE" default:"false"`
	CookieDomain string `json:"COOKIE_DOMAIN" default:""`
}

func NewConfig() (Config, error) {
	var cfg Config

	if err := envconfig.Process("AUTH", &cfg); err != nil {
		return Config{}, fmt.Errorf("proccess cookie config: %w", err)
	}

	return cfg, nil
}

func NewConfigMust() Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("get cookie config: %v", err))
	}

	return cfg
}
