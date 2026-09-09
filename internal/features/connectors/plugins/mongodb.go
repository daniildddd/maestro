package plugins

import (
	"net/url"
	"strings"

	"github.com/daniildddd/maestro/internal/core/domain"
)

type MongoDBAdapter struct{}

func (MongoDBAdapter) Match(class string) bool {
	return class == "io.debezium.connector.mongodb.MongoDbConnector"
}

func (a MongoDBAdapter) Canonicalize(config map[string]string) domain.SourceConfig {
	hosts := a.connectionHosts(config)

	cfg := domain.SourceConfig{User: a.user(config)}

	if len(hosts) > 1 {
		cfg.Hostname = strings.Join(hosts, ", ")
	} else if len(hosts) == 1 {
		cfg.Hostname, cfg.Port = splitHostPort(hosts[0])
	}

	return cfg
}

func (MongoDBAdapter) connectionHosts(config map[string]string) []string {
	conn := config["mongodb.connection.string"]
	if conn == "" {
		return nil
	}

	parsed, err := url.Parse(conn)
	if err != nil || parsed.Host == "" {
		return nil
	}

	return strings.Split(parsed.Host, ",")
}

func splitHostPort(hostport string) (host, port string) {
	host, port = hostport, ""

	if index := strings.LastIndexByte(hostport, ':'); index >= 0 {
		host, port = hostport[:index], hostport[index+1:]
	}

	return host, port
}

func (MongoDBAdapter) user(config map[string]string) string {
	if conn := config["mongodb.connection.string"]; conn != "" {
		if parsed, err := url.Parse(conn); err == nil && parsed.User != nil {
			if name := parsed.User.Username(); name != "" {
				return name
			}
		}
	}

	return config["mongodb.user"]
}
