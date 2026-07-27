package server

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr                         string        `envconfig:"ADDR" default:""`
	ReadTimeout                  time.Duration `envconfig:"READ_TIMEOUT" default:"30s"`
	ReadHeaderTimeout            time.Duration `envconfig:"READ_HEADER_TIMEOUT" default:"10s"`
	WriteTimeout                 time.Duration `envconfig:"WRITE_TIMEOUT" default:"30s"`
	IdleTimeout                  time.Duration `envconfig:"IDLE_TIMEOUT" default:"120s"`
	MaxHeaderBytes               int           `envconfig:"MAX_HEADER_BYTES" default:"8192"`
	DisableGeneralOptionsHandler bool          `envconfig:"DISABLE_GENERAL_OPTIONS_HANDLER" default:"true"`
	ShutdownTimeout              time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"30s"`
}

func NewConfigMust() Config {
	var cfg Config
	if err := envconfig.Process("SERVER", &cfg); err != nil {
		panic(fmt.Errorf("process server config: %v", err))
	}

	return cfg
}
