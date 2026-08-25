package plugins

import "github.com/daniildddd/maestro/internal/core/domain"

type Db2Adapter struct{}

func (Db2Adapter) Match(class string) bool {
	return class == "io.debezium.connector.db2.Db2Connector"
}

func (Db2Adapter) Canonicalize(config map[string]string) domain.SourceConfig {
	return domain.SourceConfig{
		Hostname: config["database.hostname"],
		Port:     config["database.port"],
		User:     config["database.user"],
		DBName:   stringPtr(config["database.dbname"]),
	}
}
