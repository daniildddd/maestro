package pgxadapter

import (
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string `envconfig:"HOST" default:"localhost"`
	Port     string `envconfig:"PORT" default:"5432"`
	User     string `envconfig:"USER" required:"true"`
	Password string `envconfig:"PASSWORD" required:"true"`
	Database string `envconfig:"DB" default:"postgres"`
	DSN      string

	Timeout time.Duration `envconfig:"TIMEOUT"`
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
	const schemeName = "postgres"

	u := &url.URL{
		Scheme: schemeName,
		User:   url.UserPassword(user, password),
		Host:   net.JoinHostPort(host, port),
		Path:   database,
	}

	u.RawQuery = "sslmode=disable"
	return u.String()
}
