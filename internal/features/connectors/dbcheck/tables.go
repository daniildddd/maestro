package dbcheck

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/daniildddd/maestro/internal/core/domain"
)

type tableScanState struct {
	problems int
	reported int
	checked  int
	regex    bool
}

func (s *tableScanState) report(
	maxChecks int,
	problem domain.ValidationCheck,
) []domain.ValidationCheck {
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
		return []domain.ValidationCheck{check(
			"postgres.table.primary_key",
			domain.CheckSeverityOK,
			"table %q has no primary key - replica identity uses a unique index",
			schema+"."+table,
		)}, false
	case "f":
		return []domain.ValidationCheck{tableCheck(
			"postgres.table.primary_key",
			domain.CheckSeverityWarning,
			schema+"."+table,
			"table %q has no primary key - REPLICA IDENTITY FULL is set (correct but WAL-heavy)",
			schema+"."+table,
		)}, false
	case "n":
		return state.report(c.cfg.MaxTableChecks, tableCheck("postgres.table.primary_key",
			domain.CheckSeverityWarning, schema+"."+table,
			"table %q has REPLICA IDENTITY NOTHING - UPDATE/DELETE events are not captured",
			schema+"."+table)), false
	default:
		return state.report(c.cfg.MaxTableChecks, fixCheck(tableCheck(
			"postgres.table.primary_key",
			domain.CheckSeverityWarning,
			schema+"."+table,
			"table %q has no primary key and default replica identity - UPDATE/DELETE events will fail",
			schema+"."+table,
		),
			"ALTER TABLE "+schema+"."+table+" ADD PRIMARY KEY (...);"+
				" or ALTER TABLE "+schema+"."+table+" REPLICA IDENTITY FULL;")), false
	}
}
