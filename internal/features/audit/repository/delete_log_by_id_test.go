package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/errs"
)

func TestDeleteLogByID(t *testing.T) {
	t.Parallel()

	const opTimeout = 100 * time.Millisecond

	id := uuid.New()

	tests := []struct {
		name   string
		setup  func(pool *MockPool)
		wantIs error
	}{
		{
			name: "success deletes existing log",
			setup: func(pool *MockPool) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Exec(mock.Anything, mock.Anything, []any{id}).
					Return(stubCommandTag{affected: 1}, nil).
					Once()
			},
		},
		{
			name: "exec error is wrapped",
			setup: func(pool *MockPool) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Exec(mock.Anything, mock.Anything, []any{id}).
					Return(nil, errs.ErrInternal).
					Once()
			},
			wantIs: errs.ErrInternal,
		},
		{
			name: "no affected rows maps to ErrAuditLogNotFound",
			setup: func(pool *MockPool) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Exec(mock.Anything, mock.Anything, []any{id}).
					Return(stubCommandTag{affected: 0}, nil).
					Once()
			},
			wantIs: errs.ErrAuditLogNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			must := require.New(t)

			pool := NewMockPool(t)
			tt.setup(pool)

			repo := newAuditRepo(pool)

			err := repo.DeleteLogByID(context.Background(), id)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
		})
	}
}
