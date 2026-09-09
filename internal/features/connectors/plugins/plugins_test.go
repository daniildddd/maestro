package plugins_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/features/connectors/plugins"
)

func TestRegistryFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		class     string
		wantType  any
		wantFound bool
	}{
		{
			name:      "postgres class resolved",
			class:     "io.debezium.connector.postgresql.PostgresConnector",
			wantType:  plugins.PostgresAdapter{},
			wantFound: true,
		},
		{
			name:      "mysql class resolved",
			class:     "io.debezium.connector.mysql.MySqlConnector",
			wantType:  plugins.MySQLAdapter{},
			wantFound: true,
		},
		{
			name:      "db2 class resolved",
			class:     "io.debezium.connector.db2.Db2Connector",
			wantType:  plugins.Db2Adapter{},
			wantFound: true,
		},
		{
			name:      "informix class resolved",
			class:     "io.debezium.connector.informix.InformixConnector",
			wantType:  plugins.InformixAdapter{},
			wantFound: true,
		},
		{
			name:      "vitess class resolved",
			class:     "io.debezium.connector.vitess.VitessConnector",
			wantType:  plugins.VitessAdapter{},
			wantFound: true,
		},
		{
			name:      "unknown class not found",
			class:     "io.debezium.connector.spanner.SpannerConnector",
			wantFound: false,
		},
		{
			name:      "empty class not found",
			class:     "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			registry := plugins.NewRegistry(
				plugins.PostgresAdapter{},
				plugins.MySQLAdapter{},
				plugins.MongoDBAdapter{},
				plugins.SQLServerAdapter{},
				plugins.OracleAdapter{},
				plugins.Db2Adapter{},
				plugins.InformixAdapter{},
				plugins.VitessAdapter{},
			)

			adapter := registry.For(tt.class)

			if !tt.wantFound {
				must.Nil(adapter)

				return
			}

			must.NotNil(adapter)
			must.IsType(tt.wantType, adapter)
		})
	}
}

func TestDb2AdapterMatch(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	adapter := plugins.Db2Adapter{}

	must.True(adapter.Match("io.debezium.connector.db2.Db2Connector"))
	must.False(adapter.Match("io.debezium.connector.postgresql.PostgresConnector"))
	must.False(adapter.Match(""))
}

func TestDb2AdapterCanonicalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config map[string]string
		want   domain.SourceConfig
	}{
		{
			name: "full db2 config",
			config: map[string]string{
				"connector.class":   "io.debezium.connector.db2.Db2Connector",
				"database.hostname": "db2-host",
				"database.port":     "50000",
				"database.user":     "dbz",
				"database.password": "********",
				"database.dbname":   "testdb",
			},
			want: domain.SourceConfig{
				Hostname: "db2-host",
				Port:     "50000",
				User:     "dbz",
				DBName:   strPtr("testdb"),
			},
		},
		{
			name: "missing dbname stays nil",
			config: map[string]string{
				"database.hostname": "db2-host",
				"database.port":     "50000",
				"database.user":     "dbz",
			},
			want: domain.SourceConfig{
				Hostname: "db2-host",
				Port:     "50000",
				User:     "dbz",
			},
		},
		{
			name:   "empty config yields zero value",
			config: map[string]string{},
			want:   domain.SourceConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			got := plugins.Db2Adapter{}.Canonicalize(tt.config)

			must.Equal(tt.want, got)
		})
	}
}

func TestInformixAdapterMatch(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	adapter := plugins.InformixAdapter{}

	must.True(adapter.Match("io.debezium.connector.informix.InformixConnector"))
	must.False(adapter.Match("io.debezium.connector.postgresql.PostgresConnector"))
	must.False(adapter.Match(""))
}

func TestInformixAdapterCanonicalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config map[string]string
		want   domain.SourceConfig
	}{
		{
			name: "full informix config",
			config: map[string]string{
				"connector.class":   "io.debezium.connector.informix.InformixConnector",
				"database.hostname": "informix-host",
				"database.port":     "50000",
				"database.user":     "dbz",
				"database.password": "********",
				"database.dbname":   "testdb",
			},
			want: domain.SourceConfig{
				Hostname: "informix-host",
				Port:     "50000",
				User:     "dbz",
				DBName:   strPtr("testdb"),
			},
		},
		{
			name: "missing dbname stays nil",
			config: map[string]string{
				"database.hostname": "informix-host",
				"database.port":     "50000",
				"database.user":     "dbz",
			},
			want: domain.SourceConfig{
				Hostname: "informix-host",
				Port:     "50000",
				User:     "dbz",
			},
		},
		{
			name:   "empty config yields zero value",
			config: map[string]string{},
			want:   domain.SourceConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			got := plugins.InformixAdapter{}.Canonicalize(tt.config)

			must.Equal(tt.want, got)
		})
	}
}

func TestMariaDBAdapterMatch(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	adapter := plugins.MariaDBAdapter{}

	must.True(adapter.Match("io.debezium.connector.mariadb.MariaDbConnector"))
	must.False(adapter.Match("io.debezium.connector.mysql.MySqlConnector"))
	must.False(adapter.Match(""))
}

func TestMariaDBAdapterCanonicalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config map[string]string
		want   domain.SourceConfig
	}{
		{
			name: "full mariadb config",
			config: map[string]string{
				"connector.class":       "io.debezium.connector.mariadb.MariaDbConnector",
				"database.hostname":     "mariadb",
				"database.port":         "3306",
				"database.user":         "debezium",
				"database.password":     "dbz",
				"database.include.list": "inventory",
			},
			want: domain.SourceConfig{
				Hostname: "mariadb",
				Port:     "3306",
				User:     "debezium",
			},
		},
		{
			name:   "empty config yields zero value",
			config: map[string]string{},
			want:   domain.SourceConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			got := plugins.MariaDBAdapter{}.Canonicalize(tt.config)

			must.Equal(tt.want, got)
		})
	}
}

func TestMySQLAdapterMatch(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	adapter := plugins.MySQLAdapter{}

	must.True(adapter.Match("io.debezium.connector.mysql.MySqlConnector"))
	must.False(adapter.Match("io.debezium.connector.postgresql.PostgresConnector"))
	must.False(adapter.Match(""))
}

func TestMySQLAdapterCanonicalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config map[string]string
		want   domain.SourceConfig
	}{
		{
			name: "full mysql config",
			config: map[string]string{
				"connector.class":       "io.debezium.connector.mysql.MySqlConnector",
				"database.hostname":     "mysql",
				"database.port":         "3306",
				"database.user":         "debezium",
				"database.password":     "dbz",
				"database.include.list": "inventory",
			},
			want: domain.SourceConfig{
				Hostname: "mysql",
				Port:     "3306",
				User:     "debezium",
			},
		},
		{
			name: "include.list is regexes, DBName stays nil even for single entry",
			config: map[string]string{
				"database.hostname":     "mysql",
				"database.port":         "3306",
				"database.user":         "debezium",
				"database.include.list": "inventory",
			},
			want: domain.SourceConfig{
				Hostname: "mysql",
				Port:     "3306",
				User:     "debezium",
			},
		},
		{
			name:   "empty config yields zero value",
			config: map[string]string{},
			want:   domain.SourceConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			got := plugins.MySQLAdapter{}.Canonicalize(tt.config)

			must.Equal(tt.want, got)
		})
	}
}

func TestOracleAdapterMatch(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	adapter := plugins.OracleAdapter{}

	must.True(adapter.Match("io.debezium.connector.oracle.OracleConnector"))
	must.False(adapter.Match("io.debezium.connector.mysql.MySqlConnector"))
	must.False(adapter.Match(""))
}

func TestOracleAdapterCanonicalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config map[string]string
		want   domain.SourceConfig
	}{
		{
			name: "full oracle config",
			config: map[string]string{
				"connector.class":   "io.debezium.connector.oracle.OracleConnector",
				"database.hostname": "oracle",
				"database.port":     "1521",
				"database.user":     "dbzuser",
				"database.dbname":   "ORCLCDB",
				"database.pdb.name": "ORCLPDB1",
			},
			want: domain.SourceConfig{
				Hostname: "oracle",
				Port:     "1521",
				User:     "dbzuser",
				DBName:   strPtr("ORCLCDB"),
			},
		},
		{
			name:   "empty config yields zero value",
			config: map[string]string{},
			want:   domain.SourceConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			got := plugins.OracleAdapter{}.Canonicalize(tt.config)

			must.Equal(tt.want, got)
		})
	}
}

func TestPostgresAdapterMatch(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	adapter := plugins.PostgresAdapter{}

	must.True(adapter.Match("io.debezium.connector.postgresql.PostgresConnector"))
	must.False(adapter.Match("io.debezium.connector.mysql.MySqlConnector"))
	must.False(adapter.Match(""))
}

func TestPostgresAdapterCanonicalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config map[string]string
		want   domain.SourceConfig
	}{
		{
			name: "full postgres config",
			config: map[string]string{
				"connector.class":    "io.debezium.connector.postgresql.PostgresConnector",
				"database.hostname":  "maestro-postgres",
				"database.port":      "5432",
				"database.user":      "postgres",
				"database.password":  "********",
				"database.dbname":    "testdb",
				"plugin.name":        "pgoutput",
				"table.include.list": "public.users",
			},
			want: domain.SourceConfig{
				Hostname:   "maestro-postgres",
				Port:       "5432",
				User:       "postgres",
				DBName:     strPtr("testdb"),
				PluginName: strPtr("pgoutput"),
			},
		},
		{
			name: "missing plugin.name stays nil",
			config: map[string]string{
				"database.hostname": "maestro-postgres",
				"database.port":     "5432",
				"database.user":     "postgres",
				"database.dbname":   "testdb",
			},
			want: domain.SourceConfig{
				Hostname: "maestro-postgres",
				Port:     "5432",
				User:     "postgres",
				DBName:   strPtr("testdb"),
			},
		},
		{
			name:   "empty config yields zero value",
			config: map[string]string{},
			want:   domain.SourceConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			got := plugins.PostgresAdapter{}.Canonicalize(tt.config)

			must.Equal(tt.want, got)
		})
	}
}

func TestSQLServerAdapterMatch(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	adapter := plugins.SQLServerAdapter{}

	must.True(adapter.Match("io.debezium.connector.sqlserver.SqlServerConnector"))
	must.False(adapter.Match("io.debezium.connector.postgresql.PostgresConnector"))
	must.False(adapter.Match(""))
}

func TestSQLServerAdapterCanonicalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config map[string]string
		want   domain.SourceConfig
	}{
		{
			name: "single database in database.names",
			config: map[string]string{
				"connector.class":   "io.debezium.connector.sqlserver.SqlServerConnector",
				"database.hostname": "sqlserver",
				"database.port":     "1433",
				"database.user":     "sa",
				"database.names":    "inventory",
			},
			want: domain.SourceConfig{
				Hostname: "sqlserver",
				Port:     "1433",
				User:     "sa",
				DBName:   strPtr("inventory"),
			},
		},
		{
			name: "multiple databases returns the list as is",
			config: map[string]string{
				"database.hostname": "sqlserver",
				"database.port":     "1433",
				"database.user":     "sa",
				"database.names":    "inventory,orders",
			},
			want: domain.SourceConfig{
				Hostname: "sqlserver",
				Port:     "1433",
				User:     "sa",
				DBName:   strPtr("inventory,orders"),
			},
		},
		{
			name:   "empty config yields zero value",
			config: map[string]string{},
			want:   domain.SourceConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			got := plugins.SQLServerAdapter{}.Canonicalize(tt.config)

			must.Equal(tt.want, got)
		})
	}
}

func TestVitessAdapterMatch(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	adapter := plugins.VitessAdapter{}

	must.True(adapter.Match("io.debezium.connector.vitess.VitessConnector"))
	must.False(adapter.Match("io.debezium.connector.postgresql.PostgresConnector"))
	must.False(adapter.Match(""))
}

func TestVitessAdapterCanonicalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config map[string]string
		want   domain.SourceConfig
	}{
		{
			name: "full vitess config",
			config: map[string]string{
				"connector.class":   "io.debezium.connector.vitess.VitessConnector",
				"database.hostname": "vitess-host",
				"database.port":     "50000",
				"database.user":     "dbz",
				"database.password": "********",
				"vitess.keyspace":   "test_keyspace",
			},
			want: domain.SourceConfig{
				Hostname: "vitess-host",
				Port:     "50000",
				User:     "dbz",
				DBName:   strPtr("test_keyspace"),
			},
		},
		{
			name: "missing keyspace stays nil",
			config: map[string]string{
				"database.hostname": "vitess-host",
				"database.port":     "50000",
				"database.user":     "dbz",
			},
			want: domain.SourceConfig{
				Hostname: "vitess-host",
				Port:     "50000",
				User:     "dbz",
			},
		},
		{
			name:   "empty config yields zero value",
			config: map[string]string{},
			want:   domain.SourceConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			got := plugins.VitessAdapter{}.Canonicalize(tt.config)

			must.Equal(tt.want, got)
		})
	}
}

func strPtr(value string) *string {
	return &value
}
