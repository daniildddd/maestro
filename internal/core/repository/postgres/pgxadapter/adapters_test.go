package pgxadapter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
	"github.com/daniildddd/maestro/internal/core/repository/postgres/pgxadapter"
)

type mapErrorsCase struct {
	name   string
	source error
	target error
	code   string
}

var mapErrorsTable = []mapErrorsCase{
	{
		name:   "nil error returns nil",
		source: nil,
		target: nil,
		code:   "",
	},
	{
		name:   "foreign key violation",
		source: &pgconn.PgError{Code: "23503"},
		target: core_postgres_pool.ErrViolatesForeignKey,
		code:   "23503",
	},
	{
		name:   "unique violation",
		source: &pgconn.PgError{Code: "23505"},
		target: core_postgres_pool.ErrDuplicate,
		code:   "23505",
	},
	{
		name:   "check violation",
		source: &pgconn.PgError{Code: "23514"},
		target: core_postgres_pool.ErrValidation,
		code:   "23514",
	},
	{
		name:   "deadlock detected",
		source: &pgconn.PgError{Code: "40P01"},
		target: core_postgres_pool.ErrDeadlock,
		code:   "40P01",
	},
	{
		name:   "unknown pg error code maps to ErrUnknown",
		source: &pgconn.PgError{Code: "42601"},
		target: core_postgres_pool.ErrUnknown,
		code:   "42601",
	},
}

type passthroughCase struct {
	name   string
	source error
	target error
}

var passthroughTable = []passthroughCase{
	{
		name:   "no rows maps to ErrNoRows",
		source: pgx.ErrNoRows,
		target: core_postgres_pool.ErrNoRows,
	},
	{
		name:   "context canceled returned as-is",
		source: context.Canceled,
		target: context.Canceled,
	},
	{
		name:   "context deadline exceeded returned as-is",
		source: context.DeadlineExceeded,
		target: context.DeadlineExceeded,
	},
}

type fakeRows struct {
	pgx.Rows

	err error
}

func (f fakeRows) Scan(_ ...any) error {
	return f.err
}

func (f fakeRows) Err() error {
	return f.err
}

type fakeRow struct {
	pgx.Row

	scanErr error
}

func (f fakeRow) Scan(_ ...any) error {
	return f.scanErr
}

func runMapErrorsTable(t *testing.T, wrap func(error) error) {
	t.Helper()

	for _, tt := range mapErrorsTable {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			err := wrap(tt.source)

			if tt.target != nil {
				var pgErr *pgconn.PgError

				must.Error(err)
				must.ErrorAs(err, &pgErr)
				is.Equal(tt.code, pgErr.Code)
				is.ErrorIs(err, tt.target)

				return
			}

			must.NoError(err)
		})
	}
}

func runPassthroughTable(t *testing.T, wrap func(error) error) {
	t.Helper()

	t.Run("passthrough errors not wrapped in ErrUnknown", func(t *testing.T) {
		t.Parallel()

		for _, tt := range passthroughTable {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				must := require.New(t)

				err := wrap(tt.source)

				must.Error(err)
				must.ErrorIs(err, tt.target)
				must.NotErrorIs(err, core_postgres_pool.ErrUnknown)
			})
		}
	})
}

func runPlainErrorTest(t *testing.T, wrap func(error) error) {
	t.Helper()

	t.Run("plain error maps to ErrUnknown", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		err := wrap(errors.New("boom"))

		must.Error(err)
		must.ErrorIs(err, core_postgres_pool.ErrUnknown)
	})
}

func TestPgxRows_Scan(t *testing.T) {
	t.Parallel()

	wrap := func(src error) error {
		return pgxadapter.PgxRows{Rows: fakeRows{err: src}}.Scan()
	}

	runMapErrorsTable(t, wrap)
	runPassthroughTable(t, wrap)
	runPlainErrorTest(t, wrap)
}

func TestPgxRows_Err(t *testing.T) {
	t.Parallel()

	wrap := func(src error) error {
		return pgxadapter.PgxRows{Rows: fakeRows{err: src}}.Err()
	}

	runMapErrorsTable(t, wrap)
	runPassthroughTable(t, wrap)
	runPlainErrorTest(t, wrap)
}

func TestPgxRow_Scan(t *testing.T) {
	t.Parallel()

	wrap := func(src error) error {
		return pgxadapter.PgxRow{Row: fakeRow{scanErr: src}}.Scan()
	}

	runMapErrorsTable(t, wrap)
	runPassthroughTable(t, wrap)
	runPlainErrorTest(t, wrap)
}
