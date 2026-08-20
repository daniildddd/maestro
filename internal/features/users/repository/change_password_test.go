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
	"github.com/daniildddd/maestro/internal/features/users/repository"
)

type stubCommandTag struct {
	affected int64
}

func (s stubCommandTag) RowsAffected() int64 { return s.affected }

func TestChangePassword(t *testing.T) {
	t.Parallel()

	const opTimeout = 100 * time.Millisecond

	userID := uuid.New()
	newHash := "new-hash"

	tests := []struct {
		name   string
		setup  func(pool *MockPool)
		wantIs error
	}{
		{
			name: "success updates password hash",
			setup: func(pool *MockPool) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Exec(
						mock.MatchedBy(func(ctx context.Context) bool {
							deadline, ok := ctx.Deadline()
							if !ok {
								return false
							}

							return time.Until(deadline) <= opTimeout
						}),
						mock.Anything,
						[]any{newHash, userID},
					).
					Return(stubCommandTag{affected: 1}, nil).
					Once()
			},
		},
		{
			name: "no rows maps to ErrUserNotFound",
			setup: func(pool *MockPool) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Exec(mock.Anything, mock.Anything, []any{newHash, userID}).
					Return(stubCommandTag{affected: 0}, nil).
					Once()
			},
			wantIs: errs.ErrUserNotFound,
		},
		{
			name: "exec error is wrapped",
			setup: func(pool *MockPool) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Exec(mock.Anything, mock.Anything, []any{newHash, userID}).
					Return(nil, errs.ErrInternal).
					Once()
			},
			wantIs: errs.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			pool := NewMockPool(t)
			tt.setup(pool)

			repo := repository.NewUsersRepository(pool)

			err := repo.ChangePassword(context.Background(), userID, newHash)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
		})
	}
}
