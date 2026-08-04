package hasher

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Cost int `default:"10" envconfig:"COST"`
}

func NewConfigMust() Config {
	var cfg Config

	if err := envconfig.Process("HASHER", &cfg); err != nil {
		panic(fmt.Errorf("process hasher config: %v", err))
	}

	return cfg
}
