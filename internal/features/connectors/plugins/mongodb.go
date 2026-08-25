package plugins

import (
	"net/url"

	"github.com/daniildddd/maestro/internal/core/domain"
)

type MongoDBAdapter struct{}

func (MongoDBAdapter) Match(class string) bool {
	return class == "io.debezium.connector.mongodb.MongoDbConnector"
}

func (a MongoDBAdapter) Canonicalize(config map[string]string) domain.SourceConfig {
	host, port := a.hostPort(config)

	return domain.SourceConfig{
		Hostname: host,
		Port:     port,
		User:     a.user(config),
	}
}

func (MongoDBAdapter) hostPort(config map[string]string) (host, port string) {
	if conn := config["mongodb.connection.string"]; conn != "" {
		if parsed, err := url.Parse(conn); err == nil && parsed.Host != "" {
			return splitHostPort(firstEntry(parsed.Host))
		}
	}

	return splitHostPort(firstEntry(config["mongodb.hosts"]))
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
