package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	core_postgres "github.com/daniildddd/maestro/internal/core/repository/postgres"
)

func TestGetLogByID(t *testing.T) {
	t.Parallel()

	const opTimeout = 100 * time.Millisecond

	id := uuid.New()

	tests := []struct {
		name      string
		setup     func(pool *MockPool, row *MockRow)
		wantErr   bool
		wantIs    error
		wantEvent domain.AuditEvent
	}{
		{
			name: "success scans event with all nullable fields",
			setup: func(pool *MockPool, row *MockRow) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					QueryRow(mock.Anything, mock.Anything, []any{id}).
					Return(row).
					Once()

				scanEventIntoRow(row, mustAuditRow(id))
			},
			wantEvent: mustAuditEvent(t, id),
		},
		{
			name: "no rows maps to ErrAuditLogNotFound",
			setup: func(pool *MockPool, row *MockRow) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					QueryRow(mock.Anything, mock.Anything, []any{id}).
					Return(row).
					Once()

				row.EXPECT().
					Scan(mock.Anything).
					Return(core_postgres.ErrNoRows).
					Once()
			},
			wantErr: true,
			wantIs:  errs.ErrAuditLogNotFound,
		},
		{
			name: "corrupt state json in state before is wrapped",
			setup: func(pool *MockPool, row *MockRow) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					QueryRow(mock.Anything, mock.Anything, []any{id}).
					Return(row).
					Once()

				fixture := mustAuditRow(id)
				fixture.stateBefore = []byte(`{invalid json`)

				scanEventIntoRow(row, fixture)
			},
			wantErr: true,
		},
		{
			name: "corrupt state json in state after is wrapped",
			setup: func(pool *MockPool, row *MockRow) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					QueryRow(mock.Anything, mock.Anything, []any{id}).
					Return(row).
					Once()

				fixture := mustAuditRow(id)
				fixture.stateAfter = []byte(`{invalid json`)

				scanEventIntoRow(row, fixture)
			},
			wantErr: true,
		},
		{
			name: "scan error is wrapped",
			setup: func(pool *MockPool, row *MockRow) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					QueryRow(mock.Anything, mock.Anything, []any{id}).
					Return(row).
					Once()

				row.EXPECT().
					Scan(mock.Anything).
					Return(errs.ErrInternal).
					Once()
			},
			wantErr: true,
			wantIs:  errs.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			must := require.New(t)

			pool := NewMockPool(t)
			row := NewMockRow(t)
			tt.setup(pool, row)

			repo := newAuditRepo(pool)

			got, err := repo.GetLogByID(context.Background(), id)

			if tt.wantErr {
				must.Error(err)

				if tt.wantIs != nil {
					must.ErrorIs(err, tt.wantIs)
				}

				return
			}

			must.NoError(err)
			is.Equal(tt.wantEvent, got)
		})
	}
}
