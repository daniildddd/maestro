//go:build integration

package pgxadapter_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
	"github.com/daniildddd/maestro/internal/core/repository/postgres/pgxadapter"
)

func newTestDSN(t *testing.T) string {
	t.Helper()

	ctx := context.Background()
	must := require.New(t)
	container, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	must.NoError(err)
	testcontainers.CleanupContainer(t, container)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	must.NoError(err)

	return dsn
}

func newTestPool(t *testing.T) *pgxadapter.Pool {
	t.Helper()
	must := require.New(t)
	cfg := pgxadapter.Config{DSN: newTestDSN(t), Timeout: 5 * time.Second}
	pool, err := pgxadapter.NewPool(context.Background(), cfg)
	must.NoError(err)
	t.Cleanup(pool.Close)

	return pool
}

func TestNewPool(t *testing.T) {
	t.Parallel()

	unreachableDSN := "postgres://invalid:invalid@localhost:1/postgres?sslmode=disable" //nolint:gosec // placeholder credentials in test fixture DSN

	tests := []struct {
		name    string
		dsn     string
		timeout time.Duration
		errMsg  string
	}{
		{
			name:    "valid DSN returns pool with configured timeout",
			dsn:     newTestDSN(t),
			timeout: 5 * time.Second,
			errMsg:  "",
		},
		{
			name:   "invalid DSN fails at parse",
			dsn:    "not-a-dsn",
			errMsg: "parse",
		},
		{
			name:   "unreachable host lazy pool fails at ping",
			dsn:    unreachableDSN,
			errMsg: "ping",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			pool, err := pgxadapter.NewPool(context.Background(), pgxadapter.Config{
				DSN:     tt.dsn,
				Timeout: tt.timeout,
			})

			if tt.errMsg != "" {
				must.Error(err)
				must.Nil(pool)
				is.ErrorContains(err, tt.errMsg)

				return
			}

			must.NoError(err)
			must.NotNil(pool)
			t.Cleanup(pool.Close)
			is.Equal(tt.timeout, pool.OpTimeout())
		})
	}
}

func TestPool_Query(t *testing.T) {
	pool := newTestPool(t)
	t.Parallel()

	tests := []struct {
		name   string
		sql    string
		target error
	}{
		{
			name:   "success returns rows",
			sql:    "SELECT 1",
			target: nil,
		},
		{
			name:   "syntax error maps to ErrUnknown",
			sql:    "SELECT FROM",
			target: core_postgres_pool.ErrUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)

			rows, err := pool.Query(context.Background(), tt.sql)

			if tt.target != nil {
				must.Error(err)
				must.Nil(rows)
				must.ErrorIs(err, tt.target)

				return
			}

			must.NoError(err)
			must.NotNil(rows)
			t.Cleanup(rows.Close)
		})
	}
}

func TestPool_Exec(t *testing.T) {
	pool := newTestPool(t)
	t.Parallel()

	tests := []struct {
		name   string
		sql    string
		target error
	}{
		{
			name:   "success returns command tag",
			sql:    "SELECT 1",
			target: nil,
		},
		{
			name:   "nonexistent table maps to ErrUnknown",
			sql:    "INSERT INTO nonexistent_table VALUES (1)",
			target: core_postgres_pool.ErrUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			tag, err := pool.Exec(context.Background(), tt.sql)

			if tt.target != nil {
				must.Error(err)
				must.Nil(tag)
				must.ErrorIs(err, tt.target)

				return
			}

			must.NoError(err)
			is.NotNil(tag)
		})
	}
}

func TestPool_QueryRow(t *testing.T) {
	pool := newTestPool(t)
	t.Parallel()

	tests := []struct {
		name   string
		sql    string
		target error
	}{
		{
			name:   "success scans value",
			sql:    "SELECT 1",
			target: nil,
		},
		{
			name:   "no rows maps to ErrNoRows",
			sql:    "SELECT 1 WHERE false",
			target: core_postgres_pool.ErrNoRows,
		},
		{
			name:   "pg error maps to ErrUnknown",
			sql:    "SELECT nextval('nonexistent_seq')",
			target: core_postgres_pool.ErrUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			var v int

			err := pool.QueryRow(context.Background(), tt.sql).Scan(&v)

			if tt.target != nil {
				must.Error(err)
				must.ErrorIs(err, tt.target)

				return
			}

			must.NoError(err)
			is.Equal(1, v)
		})
	}
}
