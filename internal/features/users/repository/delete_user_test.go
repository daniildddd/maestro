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
	"github.com/daniildddd/maestro/internal/features/users/repository"
)

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	const opTimeout = 100 * time.Millisecond

	userID := uuid.New()
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now().Add(-time.Minute)

	tests := []struct {
		name      string
		setup     func(pool *MockPool, row *MockRow)
		wantIs    error
		wantNotIs error
		wantUser  domain.User
	}{
		{
			name: "success deletes user and returns deleted state",
			setup: func(pool *MockPool, row *MockRow) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					QueryRow(
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
					Return(row).
					Once()

				scanUserIntoRow(row, mustUserRow(userID, "deleteduser", "hash", "user", createdAt, &updatedAt))
			},
			wantUser: mustNewUser(t, userID, "deleteduser", "hash", "user", createdAt, &updatedAt),
		},
		{
			name: "no rows maps to ErrUserNotFound",
			setup: func(pool *MockPool, row *MockRow) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					QueryRow(mock.Anything, mock.Anything, []any{userID}).
					Return(row).
					Once()

				row.EXPECT().
					Scan(mock.Anything).
					Return(core_postgres_pool.ErrNoRows).
					Once()
			},
			wantIs:    errs.ErrUserNotFound,
			wantNotIs: core_postgres_pool.ErrNoRows,
		},
		{
			name: "scan error is wrapped",
			setup: func(pool *MockPool, row *MockRow) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					QueryRow(mock.Anything, mock.Anything, []any{userID}).
					Return(row).
					Once()

				row.EXPECT().
					Scan(mock.Anything).
					Return(errs.ErrInternal).
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
			row := NewMockRow(t)
			tt.setup(pool, row)

			repo := repository.NewUsersRepository(pool)

			deleted, err := repo.DeleteUser(context.Background(), userID)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				if tt.wantNotIs != nil {
					must.NotErrorIs(err, tt.wantNotIs)
				}

				return
			}

			must.NoError(err)
			is.Equal(tt.wantUser, deleted)
		})
	}
}
