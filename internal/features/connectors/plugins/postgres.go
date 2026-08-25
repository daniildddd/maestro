package plugins

import "github.com/daniildddd/maestro/internal/core/domain"

type PostgresAdapter struct{}

func (PostgresAdapter) Match(class string) bool {
	return class == "io.debezium.connector.postgresql.PostgresConnector"
}

func (PostgresAdapter) Canonicalize(config map[string]string) domain.SourceConfig {
	return domain.SourceConfig{
		Hostname:   config["database.hostname"],
		Port:       config["database.port"],
		User:       config["database.user"],
		DBName:     stringPtr(config["database.dbname"]),
		PluginName: stringPtr(config["plugin.name"]),
	}
}
