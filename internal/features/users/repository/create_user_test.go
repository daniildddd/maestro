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

func TestCreateUser(t *testing.T) {
	t.Parallel()

	const opTimeout = 100 * time.Millisecond

	user := mustNewUser(t,
		uuid.New(),
		"alice",
		"hash",
		"admin",
		time.Now().Add(-time.Hour),
		nil,
	)

	tests := []struct {
		name      string
		setup     func(pool *MockPool, row *MockRow)
		wantIs    error
		wantNotIs error
		wantUser  domain.User
	}{
		{
			name: "success returns user from returning row",
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
						[]any{
							user.ID,
							user.Username,
							user.PasswordHash,
							user.Role,
							user.CreatedAt,
							user.UpdatedAt,
						},
					).
					Return(row).
					Once()

				scanUserIntoRow(row, user)
			},
			wantUser: user,
		},
		{
			name: "duplicate username maps to username conflict",
			setup: func(pool *MockPool, row *MockRow) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					QueryRow(mock.Anything, mock.Anything, mock.Anything).
					Return(row).
					Once()

				row.EXPECT().
					Scan(mock.Anything).
					Return(core_postgres_pool.ErrDuplicate).
					Once()
			},
			wantIs:    errs.ErrUsernameConflict,
			wantNotIs: core_postgres_pool.ErrDuplicate,
		},
		{
			name: "scan error is wrapped",
			setup: func(pool *MockPool, row *MockRow) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					QueryRow(mock.Anything, mock.Anything, mock.Anything).
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

			created, err := repo.CreateUser(context.Background(), user)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				if tt.wantNotIs != nil {
					must.NotErrorIs(err, tt.wantNotIs)
				}

				return
			}

			must.NoError(err)
			is.Equal(tt.wantUser, created)
		})
	}
}
