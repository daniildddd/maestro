package pgxadapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
)

type pgxRows struct {
	pgx.Rows
}

func (r pgxRows) Scan(dest ...any) error {
	err := r.Rows.Scan(dest...)
	if err != nil {
		return mapErrors(err)
	}

	return nil
}

func (r pgxRows) Err() error {
	err := r.Rows.Err()
	if err != nil {
		return mapErrors(err)
	}

	return nil
}

type pgxRow struct {
	pgx.Row
}

func (r pgxRow) Scan(dest ...any) error {
	err := r.Row.Scan(dest...)
	if err != nil {
		return mapErrors(err)
	}

	return nil
}

type pgxCommandTag struct {
	pgconn.CommandTag
}

func mapErrors(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return core_postgres_pool.ErrNoRows
	}

	const (
		pgxForeignKeyViolation = "23503"
		pgxUniqueViolation     = "23505"
		pgxCheckViolation      = "23514"
		pgxDeadlockDetected    = "40P01"
	)

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgxForeignKeyViolation:
			return fmt.Errorf(
				"%w: %v",
				core_postgres_pool.ErrViolatesForeignKey,
				err,
			)

		case pgxUniqueViolation:
			return fmt.Errorf(
				"%w: %v",
				core_postgres_pool.ErrDuplicate,
				err,
			)

		case pgxCheckViolation:
			return fmt.Errorf(
				"%w: %v",
				core_postgres_pool.ErrValidation,
				err,
			)

		case pgxDeadlockDetected:
			return fmt.Errorf(
				"%w: %v",
				core_postgres_pool.ErrDeadlock,
				err,
			)
		}
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}

	return fmt.Errorf("%w: %v", core_postgres_pool.ErrUnknown, err)
}
