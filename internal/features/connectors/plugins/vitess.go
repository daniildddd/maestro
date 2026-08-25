package plugins

import "github.com/daniildddd/maestro/internal/core/domain"

type VitessAdapter struct{}

func (VitessAdapter) Match(class string) bool {
	return class == "io.debezium.connector.vitess.VitessConnector"
}

func (VitessAdapter) Canonicalize(config map[string]string) domain.SourceConfig {
	return domain.SourceConfig{
		Hostname: config["database.hostname"],
		Port:     config["database.port"],
		User:     config["database.user"],
		DBName:   stringPtr(config["vitess.keyspace"]),
	}
}
