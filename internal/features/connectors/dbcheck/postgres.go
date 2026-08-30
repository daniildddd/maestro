package dbcheck

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

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

	steps = append(steps,
		connectionStep,
		c.permissions(ctx, conn, config),
		c.cdc(ctx, conn, config),
		c.tables(ctx, conn, config),
	)

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

func (c *PostgresChecker) cdc(
	ctx context.Context,
	conn *pgx.Conn,
	raw map[string]string,
) domain.ValidationStep {
	stepCtx, cancel := context.WithTimeout(ctx, c.cfg.StepTimeout)
	defer cancel()

	checks := c.serverSettings(stepCtx, conn)
	checks = append(checks, c.publicationChecks(stepCtx, conn, raw)...)

	return domain.ValidationStep{ID: domain.StepCDC, Checks: checks}
}

func (c *PostgresChecker) serverSettings(ctx context.Context, conn *pgx.Conn) []domain.ValidationCheck {
	checks := make([]domain.ValidationCheck, 0, 5)

	walLevel, err := scanString(ctx, conn, "SELECT current_setting('wal_level')")
	if err != nil {
		return []domain.ValidationCheck{check("postgres.cdc.wal_level", domain.CheckSeveritySkipped,
			"could not inspect wal_level: %v", err)}
	}

	if walLevel == "logical" {
		checks = append(checks, check("postgres.cdc.wal_level", domain.CheckSeverityOK,
			"wal_level = 'logical'"))
	} else {
		checks = append(checks, fixCheck(check("postgres.cdc.wal_level", domain.CheckSeverityError,
			"wal_level = %q - logical decoding requires 'logical'", walLevel),
			"ALTER SYSTEM SET wal_level = logical; then restart PostgreSQL."))
	}

	keepSize, err := scanString(ctx, conn,
		"SELECT setting FROM pg_settings WHERE name = 'max_slot_wal_keep_size'")
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		checks = append(checks, check("postgres.cdc.wal_retention", domain.CheckSeverityOK,
			"max_slot_wal_keep_size is not available on this PostgreSQL version"))
	case err != nil:
		checks = append(checks, check("postgres.cdc.wal_retention", domain.CheckSeveritySkipped,
			"could not inspect max_slot_wal_keep_size: %v", err))
	case keepSize == "-1":
		checks = append(checks, check("postgres.cdc.wal_retention", domain.CheckSeverityOK,
			"max_slot_wal_keep_size is unlimited"))
	default:
		checks = append(checks, fixCheck(check("postgres.cdc.wal_retention", domain.CheckSeverityWarning,
			"max_slot_wal_keep_size=%s - a lagging replication slot can be invalidated by checkpoints", keepSize),
			"Consider max_slot_wal_keep_size = -1 (unlimited) for CDC sources."))
	}

	usedSlots, err := scanInt(ctx, conn, "SELECT count(*) FROM pg_replication_slots")
	if err != nil {
		return append(checks, check("postgres.cdc.replication_slots", domain.CheckSeveritySkipped,
			"could not inspect replication slots: %v", err))
	}

	maxSlots, err := scanInt(ctx, conn,
		"SELECT setting::int FROM pg_settings WHERE name = 'max_replication_slots'")
	if err != nil {
		return append(checks, check("postgres.cdc.replication_slots", domain.CheckSeveritySkipped,
			"could not inspect max_replication_slots: %v", err))
	}

	if freeSlots := maxSlots - usedSlots; freeSlots < 1 {
		checks = append(checks, fixCheck(check("postgres.cdc.replication_slots",
			domain.CheckSeverityError, "no free replication slots (%d used of %d)", usedSlots, maxSlots),
			"ALTER SYSTEM SET max_replication_slots = N; then restart PostgreSQL."))
	} else {
		checks = append(checks, check("postgres.cdc.replication_slots", domain.CheckSeverityOK,
			"%d of %d replication slots free", freeSlots, maxSlots))
	}

	activeSenders, err := scanInt(ctx, conn, "SELECT count(*) FROM pg_stat_replication")
	if err != nil {
		return append(checks, check("postgres.cdc.wal_senders", domain.CheckSeveritySkipped,
			"could not inspect wal senders: %v", err))
	}

	maxSenders, err := scanInt(ctx, conn,
		"SELECT setting::int FROM pg_settings WHERE name = 'max_wal_senders'")
	if err != nil {
		return append(checks, check("postgres.cdc.wal_senders", domain.CheckSeveritySkipped,
			"could not inspect max_wal_senders: %v", err))
	}

	if freeSenders := maxSenders - activeSenders; freeSenders < 1 {
		checks = append(checks, fixCheck(check("postgres.cdc.wal_senders",
			domain.CheckSeverityError, "no free wal senders (%d active of %d)", activeSenders, maxSenders),
			"ALTER SYSTEM SET max_wal_senders = N; then restart PostgreSQL."))
	} else {
		checks = append(checks, check("postgres.cdc.wal_senders", domain.CheckSeverityOK,
			"%d of %d wal senders free", freeSenders, maxSenders))
	}

	return checks
}

func (c *PostgresChecker) publicationChecks(
	ctx context.Context,
	conn *pgx.Conn,
	raw map[string]string,
) []domain.ValidationCheck {
	pubName := publicationName(raw)

	pubCount, err := scanInt(ctx, conn,
		"SELECT count(*) FROM pg_publication WHERE pubname = $1", pubName)
	if err != nil {
		return []domain.ValidationCheck{check("postgres.cdc.publication", domain.CheckSeveritySkipped,
			"could not inspect the publication: %v", err)}
	}

	var checks []domain.ValidationCheck

	switch {
	case pubCount > 0:
		allTables, err := scanBool(ctx, conn,
			"SELECT puballtables FROM pg_publication WHERE pubname = $1", pubName)
		if err != nil {
			return append(checks, check("postgres.cdc.publication", domain.CheckSeveritySkipped,
				"could not inspect the publication: %v", err))
		}

		if allTables {
			checks = append(checks, check("postgres.cdc.publication", domain.CheckSeverityOK,
				"publication %q exists (all tables)", pubName))
		} else {
			checks = append(checks,
				check("postgres.cdc.publication", domain.CheckSeverityOK,
					"publication %q exists", pubName),
				check("postgres.cdc.puballtables", domain.CheckSeverityWarning,
					"publication %q does not cover future tables - new tables must be added manually", pubName),
			)
		}
	case publicationAutocreate(raw) == autocreateModeDisabled:
		checks = append(checks, fixCheck(fieldCheck("postgres.cdc.publication",
			domain.CheckSeverityError, "publication.name",
			"publication %q does not exist and publication.autocreate.mode is disabled", pubName),
			"CREATE PUBLICATION "+pubName+" FOR ALL TABLES;"))
	default:
		checks = append(checks, fieldCheck("postgres.cdc.publication",
			domain.CheckSeverityWarning, "publication.name",
			"publication %q does not exist - it will be auto-created", pubName))
	}

	return checks
}

type tableScanState struct {
	problems int
	reported int
	checked  int
	regex    bool
}

func (s *tableScanState) report(maxChecks int, problem domain.ValidationCheck) []domain.ValidationCheck {
	s.problems++

	if s.reported >= maxChecks {
		return nil
	}

	s.reported++

	return []domain.ValidationCheck{problem}
}

func (c *PostgresChecker) tables(
	ctx context.Context,
	conn *pgx.Conn,
	raw map[string]string,
) domain.ValidationStep {
	stepCtx, cancel := context.WithTimeout(ctx, c.cfg.StepTimeout)
	defer cancel()

	entries := splitList(raw["table.include.list"])
	if len(entries) == 0 {
		return domain.ValidationStep{ID: domain.StepTables, Checks: []domain.ValidationCheck{
			check("postgres.table.capture", domain.CheckSeverityWarning,
				"table.include.list is empty - all tables of the database will be captured"),
		}}
	}

	pubName := publicationName(raw)

	pubCount, err := scanInt(stepCtx, conn,
		"SELECT count(*) FROM pg_publication WHERE pubname = $1", pubName)
	if err != nil {
		return domain.ValidationStep{ID: domain.StepTables, Checks: []domain.ValidationCheck{
			check("postgres.table.inspect", domain.CheckSeveritySkipped,
				"could not inspect the publication: %v", err),
		}}
	}

	pub := publicationInfo{
		name:       pubName,
		exists:     pubCount > 0,
		autocreate: publicationAutocreate(raw),
	}

	var checks []domain.ValidationCheck

	state := &tableScanState{}

	for _, entry := range entries {
		if isRegexFilter(entry) {
			state.regex = true

			continue
		}

		state.checked++

		entryChecks, stop := c.checkTable(stepCtx, conn, entry, pub, state)
		checks = append(checks, entryChecks...)

		if stop {
			return domain.ValidationStep{ID: domain.StepTables, Checks: checks}
		}
	}

	if state.regex {
		checks = append(checks, check("postgres.table.regex", domain.CheckSeveritySkipped,
			"regex table filters are not resolved - matching tables were not checked"))
	}

	if state.problems > state.reported {
		checks = append(checks, check("postgres.table.problems", domain.CheckSeverityWarning,
			"%d more problem(s) in the remaining tables (truncated at %d)",
			state.problems-state.reported, c.cfg.MaxTableChecks))
	} else if state.problems == 0 && state.checked > 0 {
		checks = append(checks, check("postgres.table.readiness", domain.CheckSeverityOK,
			"%d captured tables are ready for CDC", state.checked))
	}

	return domain.ValidationStep{ID: domain.StepTables, Checks: checks}
}

type publicationInfo struct {
	name       string
	exists     bool
	autocreate string
}

func (c *PostgresChecker) checkTable(
	ctx context.Context,
	conn *pgx.Conn,
	entry string,
	pub publicationInfo,
	state *tableScanState,
) ([]domain.ValidationCheck, bool) {
	var checks []domain.ValidationCheck

	schema, table := splitSchemaTable(entry)

	exists, err := scanBool(ctx, conn,
		"SELECT to_regclass($1) IS NOT NULL", schema+"."+table)
	if err != nil {
		return []domain.ValidationCheck{check("postgres.table.exists", domain.CheckSeveritySkipped,
			"could not inspect table %q: %v", schema+"."+table, err)}, true
	}

	if !exists {
		return state.report(c.cfg.MaxTableChecks, tableCheck("postgres.table.exists",
			domain.CheckSeverityError, schema+"."+table,
			"table %q does not exist", schema+"."+table)), false
	}

	identityChecks, stop := c.replicaIdentity(ctx, conn, schema, table, state)
	if stop {
		return identityChecks, true
	}

	checks = append(checks, identityChecks...)

	persistence, err := scanString(ctx, conn,
		"SELECT relpersistence::text FROM pg_class WHERE oid = to_regclass($1)", schema+"."+table)
	if err != nil {
		return []domain.ValidationCheck{check("postgres.table.unlogged", domain.CheckSeveritySkipped,
			"could not inspect table %q: %v", schema+"."+table, err)}, true
	}

	if persistence == "u" {
		checks = append(checks, state.report(c.cfg.MaxTableChecks, fixCheck(
			tableCheck("postgres.table.unlogged", domain.CheckSeverityError, schema+"."+table,
				"table %q is UNLOGGED - streaming events will never be produced", schema+"."+table),
			"ALTER TABLE "+schema+"."+table+" SET LOGGED;"))...)
	}

	if !pub.exists {
		return checks, false
	}

	inPublication, err := scanInt(ctx, conn, `
		SELECT count(*)
		FROM pg_publication_tables
		WHERE pubname = $1 AND schemaname = $2 AND tablename = $3`,
		pub.name, schema, table)
	if err != nil {
		return []domain.ValidationCheck{check("postgres.table.publication", domain.CheckSeveritySkipped,
			"could not inspect publication membership: %v", err)}, true
	}

	if inPublication == 0 {
		severity := domain.CheckSeverityWarning
		message := "table %q is not in publication %q - it will be added automatically"
		fix := ""

		if pub.autocreate == autocreateModeDisabled {
			severity = domain.CheckSeverityError
			message = "table %q is not in publication %q and publication.autocreate.mode is disabled"
			fix = "ALTER PUBLICATION " + pub.name + " ADD TABLE " + schema + "." + table + ";"
		}

		checks = append(checks, state.report(c.cfg.MaxTableChecks, fixCheck(
			tableCheck("postgres.table.publication", severity, schema+"."+table,
				message, schema+"."+table, pub.name),
			fix))...)
	}

	return checks, false
}

func (c *PostgresChecker) replicaIdentity(
	ctx context.Context,
	conn *pgx.Conn,
	schema, table string,
	state *tableScanState,
) ([]domain.ValidationCheck, bool) {
	hasPK, err := scanBool(ctx, conn, `
		SELECT EXISTS (
			SELECT 1
			FROM pg_constraint con
			JOIN pg_class cls ON cls.oid = con.conrelid
			JOIN pg_namespace ns ON ns.oid = cls.relnamespace
			WHERE ns.nspname = $1 AND cls.relname = $2 AND con.contype = 'p'
		)`, schema, table)
	if err != nil {
		return []domain.ValidationCheck{check("postgres.table.primary_key", domain.CheckSeveritySkipped,
			"could not inspect table %q: %v", schema+"."+table, err)}, true
	}

	if hasPK {
		return []domain.ValidationCheck{check("postgres.table.primary_key", domain.CheckSeverityOK,
			"table %q has a primary key", schema+"."+table)}, false
	}

	replident, err := scanString(ctx, conn,
		"SELECT relreplident::text FROM pg_class WHERE oid = to_regclass($1)", schema+"."+table)
	if err != nil {
		return []domain.ValidationCheck{check("postgres.table.primary_key", domain.CheckSeveritySkipped,
			"could not inspect replica identity of table %q: %v", schema+"."+table, err)}, true
	}

	switch replident {
	case "i":
		return []domain.ValidationCheck{check("postgres.table.primary_key", domain.CheckSeverityOK,
			"table %q has no primary key - replica identity uses a unique index", schema+"."+table)}, false
	case "f":
		return []domain.ValidationCheck{tableCheck("postgres.table.primary_key", domain.CheckSeverityWarning,
			schema+"."+table,
			"table %q has no primary key - REPLICA IDENTITY FULL is set (correct but WAL-heavy)",
			schema+"."+table)}, false
	case "n":
		return state.report(c.cfg.MaxTableChecks, tableCheck("postgres.table.primary_key",
			domain.CheckSeverityWarning, schema+"."+table,
			"table %q has REPLICA IDENTITY NOTHING - UPDATE/DELETE events are not captured",
			schema+"."+table)), false
	default:
		return state.report(c.cfg.MaxTableChecks, fixCheck(tableCheck("postgres.table.primary_key",
			domain.CheckSeverityWarning, schema+"."+table,
			"table %q has no primary key and default replica identity - UPDATE/DELETE events will fail",
			schema+"."+table),
			"ALTER TABLE "+schema+"."+table+" ADD PRIMARY KEY (...); or ALTER TABLE "+schema+"."+table+" REPLICA IDENTITY FULL;")), false
	}
}
