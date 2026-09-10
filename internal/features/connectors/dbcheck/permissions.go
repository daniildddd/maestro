package dbcheck

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func pluginName(raw map[string]string) string {
	if name := raw["plugin.name"]; name != "" {
		return name
	}

	return pgOutputPlugin
}

func (c *PostgresChecker) permissions(
	ctx context.Context,
	conn *pgx.Conn,
	raw map[string]string,
) domain.ValidationStep {
	stepCtx, cancel := context.WithTimeout(ctx, c.cfg.StepTimeout)
	defer cancel()

	checks := make([]domain.ValidationCheck, 0, maxPermissionsChecks)

	replication, err := scanBool(stepCtx, conn,
		"SELECT rolreplication FROM pg_roles WHERE rolname = current_user")
	if err != nil {
		return domain.ValidationStep{ID: domain.StepPermissions, Checks: []domain.ValidationCheck{
			check("postgres.privilege.replication", domain.CheckSeveritySkipped,
				"could not inspect the REPLICATION attribute: %v", err),
		}}
	}

	if replication {
		checks = append(checks, check("postgres.privilege.replication", domain.CheckSeverityOK,
			"REPLICATION attribute is set"))
	} else {
		checks = append(checks, fixCheck(fieldCheck("postgres.privilege.replication",
			domain.CheckSeverityError, "database.user",
			"user %q has no REPLICATION attribute - logical decoding will fail", raw["database.user"]),
			"ALTER ROLE "+raw["database.user"]+" REPLICATION;"))
	}

	canConnect, err := scanBool(stepCtx, conn,
		"SELECT has_database_privilege(current_user, current_database(), 'CONNECT')")
	if err != nil {
		checks = append(checks, check("postgres.privilege.connect", domain.CheckSeveritySkipped,
			"could not inspect CONNECT privilege: %v", err))

		return domain.ValidationStep{ID: domain.StepPermissions, Checks: checks}
	}

	if canConnect {
		checks = append(checks, check("postgres.privilege.connect", domain.CheckSeverityOK,
			"CONNECT on %q is granted", raw["database.dbname"]))
	} else {
		checks = append(checks, fixCheck(fieldCheck("postgres.privilege.connect",
			domain.CheckSeverityError, "database.user",
			"user %q has no CONNECT privilege on %q", raw["database.user"], raw["database.dbname"]),
			"GRANT CONNECT ON DATABASE "+raw["database.dbname"]+" TO "+raw["database.user"]+";"))
	}

	if publicationAutocreate(raw) != autocreateModeDisabled {
		canCreate, err := scanBool(stepCtx, conn,
			"SELECT has_database_privilege(current_user, current_database(), 'CREATE')")
		if err != nil {
			checks = append(checks, check("postgres.privilege.create", domain.CheckSeveritySkipped,
				"could not inspect CREATE privilege: %v", err))

			return domain.ValidationStep{ID: domain.StepPermissions, Checks: checks}
		}

		if canCreate {
			checks = append(checks, check("postgres.privilege.create", domain.CheckSeverityOK,
				"CREATE on %q is granted (publication auto-creation)", raw["database.dbname"]))
		} else {
			checks = append(checks, fixCheck(fieldCheck("postgres.privilege.create",
				domain.CheckSeverityWarning, "database.user",
				"user %q has no CREATE privilege - publication auto-creation will fail", raw["database.user"]),
				"GRANT CREATE ON DATABASE "+raw["database.dbname"]+" TO "+raw["database.user"]+";"))
		}
	}

	checks = append(checks, c.schemaPrivileges(stepCtx, conn, raw)...)
	checks = append(checks, c.selectPrivileges(stepCtx, conn, raw)...)
	checks = append(checks, c.pluginChecks(stepCtx, conn, raw)...)

	return domain.ValidationStep{ID: domain.StepPermissions, Checks: checks}
}

func (c *PostgresChecker) schemaPrivileges(
	ctx context.Context,
	conn *pgx.Conn,
	raw map[string]string,
) []domain.ValidationCheck {
	seen := make(map[string]struct{})
	inspected := 0

	var checks []domain.ValidationCheck

	for _, entry := range splitList(raw["table.include.list"]) {
		if isRegexFilter(entry) {
			continue
		}

		schema, _ := splitSchemaTable(entry)

		if _, ok := seen[schema]; ok {
			continue
		}

		seen[schema] = struct{}{}
		inspected++

		usage, err := scanBool(ctx, conn,
			"SELECT has_schema_privilege(current_user, $1, 'USAGE')", schema)
		if err != nil {
			return append(checks, check("postgres.privilege.schema_usage", domain.CheckSeveritySkipped,
				"could not inspect schema %q: %v", schema, err))
		}

		if !usage {
			checks = append(checks, fixCheck(fieldCheck("postgres.privilege.schema_usage",
				domain.CheckSeverityError, "table.include.list",
				"user %q has no USAGE privilege on schema %q", raw["database.user"], schema),
				"GRANT USAGE ON SCHEMA "+schema+" TO "+raw["database.user"]+";"))
		}
	}

	if inspected > 0 && len(checks) == 0 {
		checks = append(checks, check("postgres.privilege.schema_usage", domain.CheckSeverityOK,
			"USAGE is granted on %d captured schema(s)", inspected))
	}

	return checks
}

func (c *PostgresChecker) selectPrivileges(
	ctx context.Context,
	conn *pgx.Conn,
	raw map[string]string,
) []domain.ValidationCheck {
	inspected := 0
	missing := 0
	grantedOK := 0

	var checks []domain.ValidationCheck

	for _, entry := range splitList(raw["table.include.list"]) {
		if isRegexFilter(entry) {
			continue
		}

		schema, table := splitSchemaTable(entry)
		inspected++

		var granted pgtype.Bool

		err := conn.QueryRow(ctx,
			"SELECT has_table_privilege(current_user, to_regclass($1), 'SELECT')",
			schema+"."+table,
		).Scan(&granted)
		if err != nil {
			return append(checks, check("postgres.privilege.select", domain.CheckSeveritySkipped,
				"could not inspect SELECT privilege: %v", err))
		}

		switch {
		case !granted.Valid:
			missing++
		case granted.Bool:
			grantedOK++
		default:
			checks = append(checks, fixCheck(tableCheck("postgres.privilege.select",
				domain.CheckSeverityError, schema+"."+table,
				"user %q has no SELECT privilege on %q", raw["database.user"], schema+"."+table),
				"GRANT SELECT ON TABLE "+schema+"."+table+" TO "+raw["database.user"]+";"))
		}
	}

	switch {
	case len(checks) > 0:
	case inspected == 0:
		checks = append(checks, check("postgres.privilege.select", domain.CheckSeveritySkipped,
			"table.include.list has no plain table filters to inspect"))
	case missing == inspected:
		checks = append(checks, check("postgres.privilege.select", domain.CheckSeveritySkipped,
			"captured tables do not exist - existence is reported in the tables step"))
	default:
		checks = append(checks, check("postgres.privilege.select", domain.CheckSeverityOK,
			"SELECT is granted on %d/%d existing captured tables", grantedOK, inspected-missing))
	}

	return checks
}

func (c *PostgresChecker) pluginChecks(
	ctx context.Context,
	conn *pgx.Conn,
	raw map[string]string,
) []domain.ValidationCheck {
	switch plugin := pluginName(raw); plugin {
	case pgOutputPlugin:
		return []domain.ValidationCheck{check("postgres.privilege.plugin", domain.CheckSeverityOK,
			"logical decoding plugin: pgoutput (built-in)")}
	case "wal2json":
		installed, err := scanInt(ctx, conn,
			"SELECT count(*) FROM pg_extension WHERE extname = 'wal2json'")
		if err != nil {
			return []domain.ValidationCheck{check("postgres.privilege.plugin", domain.CheckSeveritySkipped,
				"could not inspect the wal2json extension: %v", err)}
		}

		if installed == 0 {
			return []domain.ValidationCheck{fixCheck(fieldCheck("postgres.privilege.plugin",
				domain.CheckSeverityError, "plugin.name", "wal2json is not installed"),
				"CREATE EXTENSION wal2json;")}
		}

		return []domain.ValidationCheck{fieldCheck("postgres.privilege.plugin",
			domain.CheckSeverityWarning, "plugin.name",
			"wal2json is deprecated - prefer the built-in pgoutput")}
	case "decoderbufs":
		installed, err := scanInt(ctx, conn,
			"SELECT count(*) FROM pg_extension WHERE extname = 'decoderbufs'")
		if err != nil {
			return []domain.ValidationCheck{check("postgres.privilege.plugin", domain.CheckSeveritySkipped,
				"could not inspect the decoderbufs extension: %v", err)}
		}

		if installed == 0 {
			return []domain.ValidationCheck{fixCheck(fieldCheck("postgres.privilege.plugin",
				domain.CheckSeverityError, "plugin.name", "decoderbufs is not installed"),
				"CREATE EXTENSION decoderbufs;")}
		}

		return []domain.ValidationCheck{fieldCheck("postgres.privilege.plugin",
			domain.CheckSeverityWarning, "plugin.name",
			"decoderbufs is deprecated and removed in newer Debezium - switch to pgoutput")}
	default:
		return []domain.ValidationCheck{fieldCheck("postgres.privilege.plugin",
			domain.CheckSeverityWarning, "plugin.name",
			"unknown logical decoding plugin %q", plugin)}
	}
}
