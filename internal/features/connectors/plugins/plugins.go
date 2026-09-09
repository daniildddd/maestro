package plugins

import (
	"github.com/daniildddd/maestro/internal/core/domain"
)

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

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

type InformixAdapter struct{}

func (InformixAdapter) Match(class string) bool {
	return class == "io.debezium.connector.informix.InformixConnector"
}

func (InformixAdapter) Canonicalize(config map[string]string) domain.SourceConfig {
	return domain.SourceConfig{
		Hostname: config["database.hostname"],
		Port:     config["database.port"],
		User:     config["database.user"],
		DBName:   stringPtr(config["database.dbname"]),
	}
}

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

type SQLServerAdapter struct{}

func (SQLServerAdapter) Match(class string) bool {
	return class == "io.debezium.connector.sqlserver.SqlServerConnector"
}

func (SQLServerAdapter) Canonicalize(config map[string]string) domain.SourceConfig {
	return domain.SourceConfig{
		Hostname: config["database.hostname"],
		Port:     config["database.port"],
		User:     config["database.user"],
		DBName:   stringPtr(config["database.names"]),
	}
}

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
