package plugins

import "github.com/daniildddd/maestro/internal/core/domain"

type MySQLAdapter struct{}

func (MySQLAdapter) Match(class string) bool {
	return class == "io.debezium.connector.mysql.MySqlConnector"
}

func (MySQLAdapter) Canonicalize(config map[string]string) domain.SourceConfig {
	return domain.SourceConfig{
		Hostname: config["database.hostname"],
		Port:     config["database.port"],
		User:     config["database.user"],
	}
}
