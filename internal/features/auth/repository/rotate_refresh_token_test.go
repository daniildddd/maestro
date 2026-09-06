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
	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
	"github.com/daniildddd/maestro/internal/features/auth/repository"
)

type stubCommandTag struct {
	affected int64
}

func (s stubCommandTag) RowsAffected() int64 { return s.affected }

//nolint:maintidx // table-driven test: complexity comes from per-case mock setups
func TestRotateRefreshToken(t *testing.T) {
	t.Parallel()

	const opTimeout = 100 * time.Millisecond

	oldHash := "old-hash"
	newToken := domain.NewRefreshToken(
		uuid.New(),
		uuid.New(),
		"new-hash",
		time.Now(),
		time.Now().Add(time.Hour),
	)

	tests := []struct {
		name   string
		setup  func(pool *MockPool, tx *MockTx)
		wantIs error
	}{
		{
			name: "success deletes old and inserts new within one tx",
			setup: func(pool *MockPool, tx *MockTx) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Begin(mock.Anything).
					Return(tx, nil).
					Once()

				tx.EXPECT().
					Exec(mock.Anything, mock.Anything, []any{oldHash}).
					Return(stubCommandTag{affected: 1}, nil).
					Once()

				tx.EXPECT().
					Exec(
						mock.Anything,
						mock.Anything,
						[]any{newToken.ID, newToken.UserID, newToken.TokenHash, newToken.CreatedAt, newToken.ExpiresAt},
					).
					Return(stubCommandTag{affected: 1}, nil).
					Once()

				tx.EXPECT().
					Commit(mock.Anything).
					Return(nil).
					Once()

				tx.EXPECT().
					Rollback(mock.Anything).
					Return(nil).
					Once()
			},
		},
		{
			name: "delete old token error is wrapped without retry",
			setup: func(pool *MockPool, tx *MockTx) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Begin(mock.Anything).
					Return(tx, nil).
					Once()

				tx.EXPECT().
					Exec(mock.Anything, mock.Anything, []any{oldHash}).
					Return(nil, errs.ErrInternal).
					Once()

				tx.EXPECT().
					Rollback(mock.Anything).
					Return(nil).
					Once()
			},
			wantIs: errs.ErrInternal,
		},
		{
			name: "no rows on delete maps to ErrRefreshTokenNotFound",
			setup: func(pool *MockPool, tx *MockTx) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Begin(mock.Anything).
					Return(tx, nil).
					Once()

				tx.EXPECT().
					Exec(mock.Anything, mock.Anything, []any{oldHash}).
					Return(stubCommandTag{affected: 0}, nil).
					Once()

				tx.EXPECT().
					Rollback(mock.Anything).
					Return(nil).
					Once()
			},
			wantIs: errs.ErrRefreshTokenNotFound,
		},
		{
			name: "deadlock on delete is retried to success",
			setup: func(pool *MockPool, tx *MockTx) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Times(2)

				pool.EXPECT().
					Begin(mock.Anything).
					Return(tx, nil).
					Times(2)

				tx.EXPECT().
					Exec(mock.Anything, mock.Anything, []any{oldHash}).
					Return(nil, core_postgres_pool.ErrDeadlock).
					Once()

				tx.EXPECT().
					Exec(mock.Anything, mock.Anything, []any{oldHash}).
					Return(stubCommandTag{affected: 1}, nil).
					Once()

				tx.EXPECT().
					Exec(mock.Anything, mock.Anything, mock.Anything).
					Return(stubCommandTag{affected: 1}, nil).
					Once()

				tx.EXPECT().
					Commit(mock.Anything).
					Return(nil).
					Once()

				tx.EXPECT().
					Rollback(mock.Anything).
					Return(nil).
					Times(2)
			},
		},
		{
			name: "begin error is wrapped",
			setup: func(pool *MockPool, _ *MockTx) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Begin(mock.Anything).
					Return(nil, errs.ErrInternal).
					Once()
			},
			wantIs: errs.ErrInternal,
		},
		{
			name: "insert error rolls back and is wrapped",
			setup: func(pool *MockPool, tx *MockTx) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Begin(mock.Anything).
					Return(tx, nil).
					Once()

				tx.EXPECT().
					Exec(mock.Anything, mock.Anything, []any{oldHash}).
					Return(stubCommandTag{affected: 1}, nil).
					Once()

				tx.EXPECT().
					Exec(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errs.ErrInternal).
					Once()

				tx.EXPECT().
					Rollback(mock.Anything).
					Return(nil).
					Once()
			},
			wantIs: errs.ErrInternal,
		},
		{
			name: "commit error rolls back and is wrapped",
			setup: func(pool *MockPool, tx *MockTx) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Begin(mock.Anything).
					Return(tx, nil).
					Once()

				tx.EXPECT().
					Exec(mock.Anything, mock.Anything, []any{oldHash}).
					Return(stubCommandTag{affected: 1}, nil).
					Once()

				tx.EXPECT().
					Exec(mock.Anything, mock.Anything, mock.Anything).
					Return(stubCommandTag{affected: 1}, nil).
					Once()

				tx.EXPECT().
					Commit(mock.Anything).
					Return(errs.ErrInternal).
					Once()

				tx.EXPECT().
					Rollback(mock.Anything).
					Return(nil).
					Once()
			},
			wantIs: errs.ErrInternal,
		},
		{
			name: "deadlock on every attempt exhausts retries",
			setup: func(pool *MockPool, tx *MockTx) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Times(3)

				pool.EXPECT().
					Begin(mock.Anything).
					Return(tx, nil).
					Times(3)

				tx.EXPECT().
					Exec(mock.Anything, mock.Anything, []any{oldHash}).
					Return(nil, core_postgres_pool.ErrDeadlock).
					Times(3)

				tx.EXPECT().
					Rollback(mock.Anything).
					Return(nil).
					Times(3)
			},
			wantIs: core_postgres_pool.ErrDeadlock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			must := require.New(t)

			pool := NewMockPool(t)
			tx := NewMockTx(t)
			tt.setup(pool, tx)

			repo := repository.NewAuthRepository(pool)

			err := repo.RotateRefreshToken(context.Background(), oldHash, newToken)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
		})
	}
}
