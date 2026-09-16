package dbcheck

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func publicationName(raw map[string]string) string {
	if name := raw["publication.name"]; name != "" {
		return name
	}

	if dbname := raw[fieldDatabaseDBName]; dbname != "" {
		return dbname
	}

	return raw[fieldDatabaseUser]
}

func publicationAutocreate(raw map[string]string) string {
	if mode := raw["publication.autocreate.mode"]; mode != "" {
		return mode
	}

	return "all_tables"
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

func (c *PostgresChecker) serverSettings(
	ctx context.Context,
	conn *pgx.Conn,
) []domain.ValidationCheck {
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
		checks = append(checks, check(
			"postgres.cdc.wal_retention",
			domain.CheckSeverityOK,
			"max_slot_wal_keep_size is unlimited",
		))
	default:
		checks = append(checks, fixCheck(check(
			"postgres.cdc.wal_retention",
			domain.CheckSeverityWarning,
			"max_slot_wal_keep_size=%s - a lagging replication slot can be invalidated by checkpoints",
			keepSize,
		),
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
