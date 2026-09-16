package metrics

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Buckets []float64 `default:"0.05,0.1,0.25,0.5,1,2.5,5,10" envconfig:"BUCKETS"`
	Addr    string    `default:":9100"                        envconfig:"ADDR"`

	ReadHeaderTimeout time.Duration `default:"10s" envconfig:"READ_HEADER_TIMEOUT"`
	ShutdownTimeout   time.Duration `default:"10s" envconfig:"SHUTDOWN_TIMEOUT"`
}

func NewConfigMust() Config {
	var cfg Config

	if err := envconfig.Process("METRICS", &cfg); err != nil {
		panic(fmt.Errorf("process metrics config: %w", err))
	}

	return cfg
}
