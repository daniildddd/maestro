package logger

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Level  string `default:"DEBUG"    envconfig:"LEVEL"`
	Folder string `envconfig:"FOLDER" required:"true"`
}

func NewConfig() (Config, error) {
	const op = "core.logger.NewConfig"

	var config Config

	if err := envconfig.Process("LOGGER", &config); err != nil {
		return Config{}, fmt.Errorf("%s: process envconfig: %w", op, err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get Logger config: %w", err))
	}

	return config
}
