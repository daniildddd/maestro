package dbcheck

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func TestPostgresChecker_Match(t *testing.T) {
	t.Parallel()

	checker := NewPostgresChecker(Config{})

	require.True(t, checker.Match("io.debezium.connector.postgresql.PostgresConnector"))
	require.False(t, checker.Match("io.debezium.connector.mysql.MySqlConnector"))
	require.False(t, checker.Match(""))
}

func TestPostgresChecker_BuildConnConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		raw          map[string]string
		wantHost     string
		wantPort     uint16
		wantUser     string
		wantPassword string
		wantDatabase string
		wantProblems []configProblem
		wantFallback bool
	}{
		{
			name: "full config builds without problems",
			raw: map[string]string{
				"database.hostname": "pg-1",
				"database.port":     "5432",
				"database.user":     "debezium",
				"database.password": "secret",
				"database.dbname":   "postgres",
			},
			wantHost:     "pg-1",
			wantPort:     5432,
			wantUser:     "debezium",
			wantPassword: "secret",
			wantDatabase: "postgres",
			wantProblems: []configProblem{},
			wantFallback: true,
		},
		{
			name: "missing port and dbname fall back with warnings",
			raw: map[string]string{
				"database.hostname": "pg-1",
				"database.user":     "debezium",
				"database.password": "secret",
			},
			wantHost:     "pg-1",
			wantPort:     5432,
			wantUser:     "debezium",
			wantPassword: "secret",
			wantDatabase: "debezium",
			wantFallback: true,
			wantProblems: []configProblem{
				{
					field:    "database.port",
					message:  "database.port is not set - defaulting to 5432",
					severity: domain.CheckSeverityWarning,
				},
				{
					field: "database.dbname",
					message: "database.dbname is not set - checks will run " +
						`against the user database ("debezium")`,
					severity: domain.CheckSeverityWarning,
				},
			},
		},
		{
			name: "sslmode enables TLS",
			raw: map[string]string{
				"database.hostname": "pg-1",
				"database.port":     "5432",
				"database.user":     "debezium",
				"database.password": "secret",
				"database.dbname":   "postgres",
				"database.sslmode":  "require",
			},
			wantHost:     "pg-1",
			wantPort:     5432,
			wantUser:     "debezium",
			wantPassword: "secret",
			wantDatabase: "postgres",
			wantProblems: []configProblem{},
			wantFallback: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			connConfig, problems := NewPostgresChecker(Config{}).buildConnConfig(tt.raw)

			require.Equal(t, tt.wantProblems, problems)
			require.NotNil(t, connConfig)
			require.Equal(t, tt.wantHost, connConfig.Host)
			require.Equal(t, tt.wantPort, connConfig.Port)
			require.Equal(t, tt.wantUser, connConfig.User)
			require.Equal(t, tt.wantPassword, connConfig.Password)
			require.Equal(t, tt.wantDatabase, connConfig.Database)

			require.NotNil(t, connConfig.TLSConfig)
			require.True(t, connConfig.TLSConfig.InsecureSkipVerify)
			require.Equal(t, tt.wantFallback, len(connConfig.Fallbacks) > 0)
		})
	}
}

func TestPostgresChecker_BuildConnConfigMissingSettings(t *testing.T) {
	t.Parallel()

	_, problems := NewPostgresChecker(Config{}).buildConnConfig(map[string]string{})

	byField := make(map[string]domain.CheckSeverity, len(problems))
	for _, problem := range problems {
		byField[problem.field] = problem.severity
	}

	require.Len(t, problems, 5)
	require.Equal(t, domain.CheckSeverityError, byField["database.hostname"])
	require.Equal(t, domain.CheckSeverityError, byField["database.user"])
	require.Equal(t, domain.CheckSeverityError, byField["database.password"])
	require.Equal(t, domain.CheckSeverityWarning, byField["database.port"])
	require.Equal(t, domain.CheckSeverityWarning, byField["database.dbname"])
}

func TestFormatVersion(t *testing.T) {
	t.Parallel()

	require.Equal(t, "10.0", formatVersion(100000))
	require.Equal(t, "16.0", formatVersion(160000))
	require.Equal(t, "15.2", formatVersion(150002))
	require.Equal(t, "9.400", formatVersion(90400))
}

func TestPublicationName(t *testing.T) {
	t.Parallel()

	require.Equal(t, "custom",
		publicationName(map[string]string{
			"publication.name": "custom",
			"database.dbname":  "postgres",
			"database.user":    "debezium",
		}))
	require.Equal(t, "postgres",
		publicationName(map[string]string{
			"database.dbname": "postgres",
			"database.user":   "debezium",
		}))
	require.Equal(t, "debezium",
		publicationName(map[string]string{"database.user": "debezium"}))
	require.Empty(t, publicationName(map[string]string{}))
}

func TestPublicationAutocreate(t *testing.T) {
	t.Parallel()

	require.Equal(t, "disabled",
		publicationAutocreate(map[string]string{"publication.autocreate.mode": "disabled"}))
	require.Equal(t, "all_tables", publicationAutocreate(map[string]string{}))
}

func TestPluginName(t *testing.T) {
	t.Parallel()

	require.Equal(t, "wal2json",
		pluginName(map[string]string{"plugin.name": "wal2json"}))
	require.Equal(t, pgOutputPlugin, pluginName(map[string]string{}))
}

func TestTableScanState_Report(t *testing.T) {
	t.Parallel()

	state := &tableScanState{}

	problem := domain.ValidationCheck{ID: "postgres.table.exists"}

	require.Len(t, state.report(2, problem), 1)
	require.Len(t, state.report(2, problem), 1)
	require.Empty(t, state.report(2, problem))
	require.Equal(t, 3, state.problems)
	require.Equal(t, 2, state.reported)

	zero := &tableScanState{}
	require.Empty(t, zero.report(0, problem))
	require.Equal(t, 1, zero.problems)
	require.Equal(t, 0, zero.reported)
}

func TestPostgresChecker_CheckMissingConfigFatals(t *testing.T) {
	t.Parallel()

	steps, err := NewPostgresChecker(Config{}).Check(context.Background(), map[string]string{})

	require.NoError(t, err)
	require.Len(t, steps, 1+len(dbStepOrder)-1)

	require.Equal(t, domain.StepConnection, steps[0].ID)

	fields := make([]string, 0, len(steps[0].Checks))
	for _, gotCheck := range steps[0].Checks {
		require.Equal(t, domain.CheckSeverityError, gotCheck.Severity)
		require.Equal(t, "postgres.connection.config", gotCheck.ID)
		fields = append(fields, gotCheck.Field)
	}

	require.ElementsMatch(t,
		[]string{"database.hostname", "database.user", "database.password"},
		fields,
	)

	for i, stepID := range dbStepOrder[1:] {
		step := steps[1+i]
		require.Equal(t, stepID, step.ID)
		require.Len(t, step.Checks, 1)
		require.Equal(t, domain.CheckSeveritySkipped, step.Checks[0].Severity)
		require.Contains(t, step.Checks[0].Message, "required connection settings are missing")
	}
}

func TestPostgresChecker_CheckUnreachableHost(t *testing.T) {
	t.Parallel()

	steps, err := NewPostgresChecker(Config{StepTimeout: 5 * time.Second}).Check(
		context.Background(),
		map[string]string{
			"database.hostname": "127.0.0.1",
			"database.port":     "1",
			"database.user":     "debezium",
			"database.password": "secret",
			"database.dbname":   "postgres",
		},
	)

	require.NoError(t, err)
	require.Len(t, steps, 1+len(dbStepOrder)-1)

	require.Equal(t, domain.StepConnection, steps[0].ID)
	require.Len(t, steps[0].Checks, 1)
	require.Equal(t, domain.CheckSeverityError, steps[0].Checks[0].Severity)
	require.Equal(t, "postgres.connection.auth", steps[0].Checks[0].ID)
	require.Contains(t, steps[0].Checks[0].Message, "could not connect")

	for i, stepID := range dbStepOrder[1:] {
		step := steps[1+i]
		require.Equal(t, stepID, step.ID)
		require.Len(t, step.Checks, 1)
		require.Equal(t, domain.CheckSeveritySkipped, step.Checks[0].Severity)
		require.Contains(t, step.Checks[0].Message, "connection failed, checks skipped")
	}
}

func TestPostgresChecker_TablesEmptyList(t *testing.T) {
	t.Parallel()

	step := NewPostgresChecker(Config{}).tables(context.Background(), nil, map[string]string{})

	require.Equal(t, domain.StepTables, step.ID)
	require.Len(t, step.Checks, 1)
	require.Equal(t, domain.CheckSeverityWarning, step.Checks[0].Severity)
	require.Equal(t, "postgres.table.capture", step.Checks[0].ID)
}

func TestPostgresChecker_SchemaPrivilegesEmptyList(t *testing.T) {
	t.Parallel()

	checks := NewPostgresChecker(Config{}).schemaPrivileges(
		context.Background(),
		nil,
		map[string]string{},
	)

	require.Empty(t, checks)
}

func TestPostgresChecker_SelectPrivilegesEmptyList(t *testing.T) {
	t.Parallel()

	checks := NewPostgresChecker(Config{}).selectPrivileges(
		context.Background(),
		nil,
		map[string]string{},
	)

	require.Len(t, checks, 1)
	require.Equal(t, domain.CheckSeveritySkipped, checks[0].Severity)
	require.Equal(t, "postgres.privilege.select", checks[0].ID)
}

func TestPostgresChecker_PluginChecksWithoutDB(t *testing.T) {
	t.Parallel()

	checker := NewPostgresChecker(Config{})

	builtin := checker.pluginChecks(context.Background(), nil,
		map[string]string{"plugin.name": "pgoutput"})
	require.Len(t, builtin, 1)
	require.Equal(t, domain.CheckSeverityOK, builtin[0].Severity)
	require.Equal(t, "postgres.privilege.plugin", builtin[0].ID)

	unknown := checker.pluginChecks(context.Background(), nil,
		map[string]string{"plugin.name": "mystery"})
	require.Len(t, unknown, 1)
	require.Equal(t, domain.CheckSeverityWarning, unknown[0].Severity)
	require.Contains(t, unknown[0].Message, "mystery")
}
