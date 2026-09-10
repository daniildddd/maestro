package dbcheck

import (
	"context"

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

	//nolint:errcheck // close error is not actionable during validation
	defer func() { _ = conn.Close(connCtx) }()

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
