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
		checks = append(checks, fixCheck(check(
			"postgres.connection.version",
			domain.CheckSeverityError,
			"PostgreSQL %s does not support pgoutput - version 10 or newer is required",
			formatVersion(version),
		),
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

func (c *PostgresChecker) buildConnConfig(
	raw map[string]string,
) (*pgx.ConnConfig, []configProblem) {
	host := raw[fieldDatabaseHostname]
	port := raw[fieldDatabasePort]
	user := raw[fieldDatabaseUser]
	password := raw[fieldDatabasePassword]
	dbname := raw[fieldDatabaseDBName]

	problems := make([]configProblem, 0, maxConfigProblems)

	if host == "" {
		problems = append(problems, configProblem{
			field:    fieldDatabaseHostname,
			message:  "database.hostname is required",
			severity: domain.CheckSeverityError,
		})
	}

	if user == "" {
		problems = append(problems, configProblem{
			field:    fieldDatabaseUser,
			message:  "database.user is required",
			severity: domain.CheckSeverityError,
		})
	}

	if password == "" {
		problems = append(problems, configProblem{
			field:    fieldDatabasePassword,
			message:  "database.password is required",
			severity: domain.CheckSeverityError,
		})
	}

	if port == "" {
		port = "5432"

		problems = append(problems, configProblem{
			field:    fieldDatabasePort,
			message:  "database.port is not set - defaulting to 5432",
			severity: domain.CheckSeverityWarning,
		})
	}

	if dbname == "" {
		dbname = user

		problems = append(problems, configProblem{
			field: fieldDatabaseDBName,
			message: fmt.Sprintf(
				"database.dbname is not set - checks will run against the user database (%q)",
				user,
			),
			severity: domain.CheckSeverityWarning,
		})
	}

	dsn := &url.URL{
		Scheme: postgresScheme,
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
			field:    fieldDatabaseHostname,
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
