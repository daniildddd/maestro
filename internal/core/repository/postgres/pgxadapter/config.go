package pgxadapter

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string `json:"HOST" default:"localhost"`
	Port     string `json:"PORT" default:"5432"`
	User     string `json:"USER" required:"true"`
	Password string `json:"PASSWORD" required:"true"`
	Database string `json:"DATABASE" default:"postgres"`
	DSN      string

	Timeout time.Duration `json:"TIMEOUT"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("POSTGRES", &config); err != nil {
		return Config{}, fmt.Errorf("process config: %w", err)
	}

	config.DSN = createDSN(
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Database,
	)

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("get database config: %v", err))
	}

	return config
}

func createDSN(host, port, user, password, database string) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, database,
	)
}
