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
		name:   "foreign key violation maps to ErrViolatesForeignKey",
		source: &pgconn.PgError{Code: "23503"},
		target: core_postgres_pool.ErrViolatesForeignKey,
		code:   "23503",
	},
	{
		name:   "unique violation maps to ErrDuplicate",
		source: &pgconn.PgError{Code: "23505"},
		target: core_postgres_pool.ErrDuplicate,
		code:   "23505",
	},
	{
		name:   "not null violation maps to ErrValidation",
		source: &pgconn.PgError{Code: "23502"},
		target: core_postgres_pool.ErrValidation,
		code:   "23502",
	},
	{
		name:   "check violation maps to ErrValidation",
		source: &pgconn.PgError{Code: "23514"},
		target: core_postgres_pool.ErrValidation,
		code:   "23514",
	},
	{
		name:   "deadlock detected maps to ErrDeadlock",
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

type stubTx struct {
	pgx.Tx

	execTag     pgconn.CommandTag
	execErr     error
	queryRow    pgx.Row
	commitErr   error
	rollbackErr error
}

func (s stubTx) Exec(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
	return s.execTag, s.execErr
}

func (s stubTx) QueryRow(_ context.Context, _ string, _ ...any) pgx.Row {
	return s.queryRow
}

func (s stubTx) Commit(_ context.Context) error {
	return s.commitErr
}

func (s stubTx) Rollback(_ context.Context) error {
	return s.rollbackErr
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

func TestPgxTx_Exec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		execErr error
		wantIs  error
		wantNot error
	}{
		{
			name:    "success returns command tag",
			execErr: nil,
		},
		{
			name:    "no rows maps to ErrNoRows",
			execErr: pgx.ErrNoRows,
			wantIs:  core_postgres_pool.ErrNoRows,
			wantNot: core_postgres_pool.ErrUnknown,
		},
		{
			name:    "plain error maps to ErrUnknown",
			execErr: errors.New("boom"),
			wantIs:  core_postgres_pool.ErrUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			tx := pgxadapter.PgxTx{Tx: stubTx{
				execTag: pgconn.NewCommandTag("SELECT 1"),
				execErr: tt.execErr,
			}}

			tag, err := tx.Exec(context.Background(), "SELECT 1")

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				if tt.wantNot != nil {
					must.NotErrorIs(err, tt.wantNot)
				}

				return
			}

			must.NoError(err)
			is.Equal(int64(1), tag.RowsAffected())
		})
	}
}

func TestPgxTx_QueryRow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		scanErr error
		wantIs  error
		wantNot error
	}{
		{
			name:    "success scans row",
			scanErr: nil,
		},
		{
			name:    "no rows maps to ErrNoRows",
			scanErr: pgx.ErrNoRows,
			wantIs:  core_postgres_pool.ErrNoRows,
			wantNot: core_postgres_pool.ErrUnknown,
		},
		{
			name:    "plain error maps to ErrUnknown",
			scanErr: errors.New("boom"),
			wantIs:  core_postgres_pool.ErrUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			tx := pgxadapter.PgxTx{Tx: stubTx{
				queryRow: fakeRow{scanErr: tt.scanErr},
			}}

			err := tx.QueryRow(context.Background(), "SELECT 1").Scan()

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				if tt.wantNot != nil {
					must.NotErrorIs(err, tt.wantNot)
				}

				return
			}

			must.NoError(err)
		})
	}
}

func TestPgxTx_Commit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		commitErr error
		wantIs    error
		wantNot   error
	}{
		{
			name:      "success returns nil",
			commitErr: nil,
		},
		{
			name:      "no rows maps to ErrNoRows",
			commitErr: pgx.ErrNoRows,
			wantIs:    core_postgres_pool.ErrNoRows,
			wantNot:   core_postgres_pool.ErrUnknown,
		},
		{
			name:      "context canceled returned as-is",
			commitErr: context.Canceled,
			wantIs:    context.Canceled,
			wantNot:   core_postgres_pool.ErrUnknown,
		},
		{
			name:      "plain error maps to ErrUnknown",
			commitErr: errors.New("boom"),
			wantIs:    core_postgres_pool.ErrUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			tx := pgxadapter.PgxTx{Tx: stubTx{commitErr: tt.commitErr}}

			err := tx.Commit(context.Background())

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				if tt.wantNot != nil {
					must.NotErrorIs(err, tt.wantNot)
				}

				return
			}

			must.NoError(err)
		})
	}
}

func TestPgxTx_Rollback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		rollbackErr error
		wantIs      error
		wantNot     error
	}{
		{
			name:        "success returns nil",
			rollbackErr: nil,
		},
		{
			name:        "no rows maps to ErrNoRows",
			rollbackErr: pgx.ErrNoRows,
			wantIs:      core_postgres_pool.ErrNoRows,
			wantNot:     core_postgres_pool.ErrUnknown,
		},
		{
			name:        "context deadline exceeded returned as-is",
			rollbackErr: context.DeadlineExceeded,
			wantIs:      context.DeadlineExceeded,
			wantNot:     core_postgres_pool.ErrUnknown,
		},
		{
			name:        "plain error maps to ErrUnknown",
			rollbackErr: errors.New("boom"),
			wantIs:      core_postgres_pool.ErrUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			tx := pgxadapter.PgxTx{Tx: stubTx{rollbackErr: tt.rollbackErr}}

			err := tx.Rollback(context.Background())

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				if tt.wantNot != nil {
					must.NotErrorIs(err, tt.wantNot)
				}

				return
			}

			must.NoError(err)
		})
	}
}
