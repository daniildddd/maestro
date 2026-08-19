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

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	const opTimeout = 100 * time.Millisecond

	userID := uuid.New()

	tests := []struct {
		name   string
		setup  func(pool *MockPool)
		wantIs error
	}{
		{
			name: "success deletes user by id within op timeout",
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
						[]any{userID},
					).
					Return(nil, nil).
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
					Exec(mock.Anything, mock.Anything, []any{userID}).
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

			err := repo.DeleteUser(context.Background(), userID)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
		})
	}
}
