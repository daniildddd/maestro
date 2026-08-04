package server

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr                         string        `default:"localhost" envconfig:"ADDR"`
	ReadTimeout                  time.Duration `default:"30s"       envconfig:"READ_TIMEOUT"`
	ReadHeaderTimeout            time.Duration `default:"10s"       envconfig:"READ_HEADER_TIMEOUT"`
	WriteTimeout                 time.Duration `default:"30s"       envconfig:"WRITE_TIMEOUT"`
	IdleTimeout                  time.Duration `default:"120s"      envconfig:"IDLE_TIMEOUT"`
	MaxHeaderBytes               int           `default:"8192"      envconfig:"MAX_HEADER_BYTES"`
	DisableGeneralOptionsHandler bool          `default:"true"      envconfig:"DISABLE_GENERAL_OPTIONS_HANDLER"`
	ShutdownTimeout              time.Duration `default:"30s"       envconfig:"SHUTDOWN_TIMEOUT"`
}

func NewConfigMust() Config {
	var cfg Config

	if err := envconfig.Process("SERVER", &cfg); err != nil {
		panic(fmt.Errorf("process server config: %w", err))
	}

	return cfg
}
