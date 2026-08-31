package pgxadapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
)

type PgxRows struct {
	pgx.Rows
}

func (r PgxRows) Scan(dest ...any) error {
	err := r.Rows.Scan(dest...)
	if err != nil {
		return mapErrors(err)
	}

	return nil
}

func (r PgxRows) Err() error {
	err := r.Rows.Err()
	if err != nil {
		return mapErrors(err)
	}

	return nil
}

type PgxRow struct {
	pgx.Row
}

func (r PgxRow) Scan(dest ...any) error {
	err := r.Row.Scan(dest...)
	if err != nil {
		return mapErrors(err)
	}

	return nil
}

type PgxCommandTag struct {
	pgconn.CommandTag
}

type PgxTx struct {
	pgx.Tx
}

func (t PgxTx) Exec(
	ctx context.Context,
	sql string,
	args ...any,
) (core_postgres_pool.CommandTag, error) {
	tag, err := t.Tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, mapErrors(err)
	}

	return PgxCommandTag{tag}, nil
}

func (t PgxTx) QueryRow(
	ctx context.Context,
	sql string,
	args ...any,
) core_postgres_pool.Row {
	return PgxRow{t.Tx.QueryRow(ctx, sql, args...)}
}

func (t PgxTx) Commit(ctx context.Context) error {
	return mapErrors(t.Tx.Commit(ctx))
}

func (t PgxTx) Rollback(ctx context.Context) error {
	return mapErrors(t.Tx.Rollback(ctx))
}

func mapErrors(err error) error {
	const op = "postgres.pgxadapter.mapErrors"

	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return core_postgres_pool.ErrNoRows
	}

	const (
		pgxNotNullViolation    = "23502"
		pgxForeignKeyViolation = "23503"
		pgxUniqueViolation     = "23505"
		pgxCheckViolation      = "23514"
		pgxDeadlockDetected    = "40P01"
	)

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgxNotNullViolation, pgxCheckViolation:
			return fmt.Errorf(
				"%s: %w: %w",
				op,
				core_postgres_pool.ErrValidation,
				err,
			)

		case pgxForeignKeyViolation:
			return fmt.Errorf(
				"%s: %w: %w",
				op,
				core_postgres_pool.ErrViolatesForeignKey,
				err,
			)

		case pgxUniqueViolation:
			return fmt.Errorf(
				"%s: %w: %w",
				op,
				core_postgres_pool.ErrDuplicate,
				err,
			)

		case pgxDeadlockDetected:
			return fmt.Errorf(
				"%s: %w: %w",
				op,
				core_postgres_pool.ErrDeadlock,
				err,
			)
		}
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}

	return fmt.Errorf("%s: %w: %w", op, core_postgres_pool.ErrUnknown, err)
}
