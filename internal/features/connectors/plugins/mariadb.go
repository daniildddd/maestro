package plugins

import "github.com/daniildddd/maestro/internal/core/domain"

type MariaDBAdapter struct{}

func (MariaDBAdapter) Match(class string) bool {
	return class == "io.debezium.connector.mariadb.MariaDbConnector"
}

func (MariaDBAdapter) Canonicalize(config map[string]string) domain.SourceConfig {
	return domain.SourceConfig{
		Hostname: config["database.hostname"],
		Port:     config["database.port"],
		User:     config["database.user"],
	}
}
