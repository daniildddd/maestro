package dbcheck

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/jackc/pgx/v5"

	"github.com/daniildddd/maestro/internal/core/domain"
)

const (
	postgresClass          = "io.debezium.connector.postgresql.PostgresConnector"
	pgOutputPlugin         = "pgoutput"
	autocreateModeDisabled = "disabled"
	maxConfigProblems      = 6
	maxPermissionsChecks   = 6
)

type PostgresChecker struct {
	cfg Config
}

func NewPostgresChecker(cfg Config) *PostgresChecker {
	return &PostgresChecker{cfg: cfg}
}

func (c *PostgresChecker) Match(class string) bool {
	return class == postgresClass
}

func (c *PostgresChecker) Check(
	ctx context.Context,
	config map[string]string,
) ([]domain.ValidationStep, error) {
	connConfig, problems := c.buildConnConfig(config)

	steps := make([]domain.ValidationStep, 0, len(dbStepOrder))

	fatals := make([]configProblem, 0, len(problems))
	warnings := make([]configProblem, 0, len(problems))

	for _, problem := range problems {
		if problem.severity == domain.CheckSeverityError {
			fatals = append(fatals, problem)

			continue
		}

		warnings = append(warnings, problem)
	}

	if len(fatals) > 0 {
		checks := make([]domain.ValidationCheck, 0, len(fatals))

		for _, problem := range fatals {
			checks = append(checks, fieldCheck(
				"postgres.connection.config", domain.CheckSeverityError,
				problem.field, "%s", problem.message,
			))
		}

		steps = append(steps, domain.ValidationStep{ID: domain.StepConnection, Checks: checks})
		steps = append(steps, skippedDBSteps(domain.StepConnection,
			"required connection settings are missing")...)

		return steps, nil
	}

	connCtx, cancel := context.WithTimeout(ctx, c.cfg.StepTimeout)
	defer cancel()

	conn, err := pgx.ConnectConfig(connCtx, connConfig)
	if err != nil {
		steps = append(steps, domain.ValidationStep{
			ID: domain.StepConnection,
			Checks: []domain.ValidationCheck{fieldCheck(
				"postgres.connection.auth", domain.CheckSeverityError,
				"database.hostname", "could not connect: %v", err,
			)},
		})
		steps = append(steps, skippedDBSteps(domain.StepConnection,
			"connection failed, checks skipped")...)

		return steps, nil
	}

	defer func() { _ = conn.Close(connCtx) }() //nolint:errcheck // close error is not actionable during validation

	connectionStep := c.connectionOK(connCtx, conn, connConfig, config)

	if len(warnings) > 0 {
		warningChecks := make([]domain.ValidationCheck, 0, len(warnings)+len(connectionStep.Checks))

		for _, problem := range warnings {
			warningChecks = append(warningChecks, fieldCheck(
				"postgres.connection.config", domain.CheckSeverityWarning,
				problem.field, "%s", problem.message,
			))
		}

		warningChecks = append(warningChecks, connectionStep.Checks...)
		connectionStep.Checks = warningChecks
	}

	steps = append(steps, connectionStep)
	steps = append(steps, skippedDBSteps(domain.StepConnection,
		"further checks are not implemented yet")...)

	return steps, nil
}

func (c *PostgresChecker) connectionOK(
	ctx context.Context,
	conn *pgx.Conn,
	cfg *pgx.ConnConfig,
	raw map[string]string,
) domain.ValidationStep {
	checks := []domain.ValidationCheck{
		check("postgres.connection.auth", domain.CheckSeverityOK,
			"connected as %q to %q at %s",
			cfg.User, cfg.Database, net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port))),
		),
	}

	versionText, err := scanString(ctx, conn, "SELECT current_setting('server_version_num')")
	if err != nil {
		checks = append(checks,
			check("postgres.connection.version", domain.CheckSeveritySkipped,
				"could not inspect the server version: %v", err),
			check("postgres.connection.primary", domain.CheckSeveritySkipped,
				"could not inspect the recovery state: %v", err),
		)

		return domain.ValidationStep{ID: domain.StepConnection, Checks: checks}
	}

	version, err := strconv.Atoi(versionText)
	if err != nil {
		checks = append(checks, check("postgres.connection.version", domain.CheckSeveritySkipped,
			"unexpected server_version_num %q: %v", versionText, err))

		return domain.ValidationStep{ID: domain.StepConnection, Checks: checks}
	}

	if version < 100000 && pluginName(raw) == pgOutputPlugin {
		checks = append(checks, fixCheck(check("postgres.connection.version", domain.CheckSeverityError,
			"PostgreSQL %s does not support pgoutput - version 10 or newer is required", formatVersion(version)),
			"Upgrade PostgreSQL, or install wal2json and set plugin.name=wal2json."))
	} else {
		checks = append(checks, check("postgres.connection.version", domain.CheckSeverityOK,
			"PostgreSQL %s", formatVersion(version)))
	}

	recovery, err := scanBool(ctx, conn, "SELECT pg_is_in_recovery()")
	if err != nil {
		checks = append(checks, check("postgres.connection.primary", domain.CheckSeveritySkipped,
			"could not inspect the recovery state: %v", err))

		return domain.ValidationStep{ID: domain.StepConnection, Checks: checks}
	}

	switch {
	case !recovery:
		checks = append(checks, check("postgres.connection.primary", domain.CheckSeverityOK,
			"source is a primary"))
	case recovery && version < 160000:
		checks = append(checks, check("postgres.connection.primary", domain.CheckSeverityError,
			"source is a standby - logical decoding on a standby requires PostgreSQL 16 or newer"))
	default:
		checks = append(checks, check("postgres.connection.primary", domain.CheckSeverityWarning,
			"source is a standby - logical decoding works, but the replication slot is tied to the primary"))
	}

	return domain.ValidationStep{ID: domain.StepConnection, Checks: checks}
}

func (c *PostgresChecker) buildConnConfig(raw map[string]string) (*pgx.ConnConfig, []configProblem) {
	host := raw["database.hostname"]
	port := raw["database.port"]
	user := raw["database.user"]
	password := raw["database.password"]
	dbname := raw["database.dbname"]

	problems := make([]configProblem, 0, maxConfigProblems)

	if host == "" {
		problems = append(problems, configProblem{
			field:    "database.hostname",
			message:  "database.hostname is required",
			severity: domain.CheckSeverityError,
		})
	}

	if user == "" {
		problems = append(problems, configProblem{
			field:    "database.user",
			message:  "database.user is required",
			severity: domain.CheckSeverityError,
		})
	}

	if password == "" {
		problems = append(problems, configProblem{
			field:    "database.password",
			message:  "database.password is required",
			severity: domain.CheckSeverityError,
		})
	}

	if port == "" {
		port = "5432"

		problems = append(problems, configProblem{
			field:    "database.port",
			message:  "database.port is not set - defaulting to 5432",
			severity: domain.CheckSeverityWarning,
		})
	}

	if dbname == "" {
		dbname = user

		problems = append(problems, configProblem{
			field:    "database.dbname",
			message:  fmt.Sprintf("database.dbname is not set - checks will run against the user database (%q)", user),
			severity: domain.CheckSeverityWarning,
		})
	}

	dsn := &url.URL{
		Scheme: "postgres",
		Host:   net.JoinHostPort(host, port),
		Path:   "/" + dbname,
	}

	query := dsn.Query()

	if sslmode := raw["database.sslmode"]; sslmode != "" {
		query.Set("sslmode", sslmode)
	}

	dsn.RawQuery = query.Encode()

	connConfig, err := pgx.ParseConfig(dsn.String())
	if err != nil {
		problems = append(problems, configProblem{
			field:    "database.hostname",
			message:  "could not build connection settings",
			severity: domain.CheckSeverityError,
		})

		return nil, problems
	}

	connConfig.User = user
	connConfig.Password = password

	return connConfig, problems
}

func formatVersion(version int) string {
	return fmt.Sprintf("%d.%d", version/10000, version%10000)
}

func scanString(ctx context.Context, conn *pgx.Conn, query string, args ...any) (string, error) {
	var value string

	err := conn.QueryRow(ctx, query, args...).Scan(&value)

	return value, err
}

func scanInt(ctx context.Context, conn *pgx.Conn, query string, args ...any) (int, error) {
	var value int

	err := conn.QueryRow(ctx, query, args...).Scan(&value)

	return value, err
}

func scanBool(ctx context.Context, conn *pgx.Conn, query string, args ...any) (bool, error) {
	var value bool

	err := conn.QueryRow(ctx, query, args...).Scan(&value)

	return value, err
}
func publicationName(raw map[string]string) string {
	if name := raw["publication.name"]; name != "" {
		return name
	}

	if dbname := raw["database.dbname"]; dbname != "" {
		return dbname
	}

	return raw["database.user"]
}

func publicationAutocreate(raw map[string]string) string {
	if mode := raw["publication.autocreate.mode"]; mode != "" {
		return mode
	}

	return "all_tables"
}

func pluginName(raw map[string]string) string {
	if name := raw["plugin.name"]; name != "" {
		return name
	}

	return pgOutputPlugin
}
