package pgxadapter

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
)

type Pool struct {
	*pgxpool.Pool

	opTimeout time.Duration
}

func NewPool(
	ctx context.Context,
	config Config,
) (*Pool, error) {
	const op = "postgres.pgxadapter.NewPool"

	pgxConfig, err := pgxpool.ParseConfig(config.DSN)
	if err != nil {
		return nil, fmt.Errorf("%s: parse pgxconfig: %w", op, err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("%s: create pgxpool: %w", op, err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()

		return nil, fmt.Errorf("%s: pgxpool ping: %w", op, err)
	}

	return &Pool{
		Pool:      pool,
		opTimeout: config.Timeout,
	}, nil
}

func (p *Pool) Query(
	ctx context.Context,
	sql string,
	args ...any,
) (core_postgres_pool.Rows, error) {
	rows, err := p.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, mapErrors(err)
	}

	return PgxRows{rows}, nil
}

func (p *Pool) QueryRow(
	ctx context.Context,
	sql string,
	args ...any,
) core_postgres_pool.Row {
	row := p.Pool.QueryRow(ctx, sql, args...)

	return PgxRow{row}
}

func (p *Pool) Exec(
	ctx context.Context,
	sql string,
	args ...any,
) (core_postgres_pool.CommandTag, error) {
	commandTag, err := p.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, mapErrors(err)
	}

	return PgxCommandTag{commandTag}, nil
}

func (p *Pool) OpTimeout() time.Duration {
	return p.opTimeout
}
