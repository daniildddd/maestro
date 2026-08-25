package plugins

import "github.com/daniildddd/maestro/internal/core/domain"

type OracleAdapter struct{}

func (OracleAdapter) Match(class string) bool {
	return class == "io.debezium.connector.oracle.OracleConnector"
}

func (OracleAdapter) Canonicalize(config map[string]string) domain.SourceConfig {
	return domain.SourceConfig{
		Hostname: config["database.hostname"],
		Port:     config["database.port"],
		User:     config["database.user"],
		DBName:   stringPtr(config["database.dbname"]),
	}
}
