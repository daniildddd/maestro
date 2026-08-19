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

//nolint:gocognit,cyclop,revive,maintidx // table-driven test: complexity comes from mock setup Run blocks
func TestGetUserByID(t *testing.T) {
	t.Parallel()

	const opTimeout = 100 * time.Millisecond

	id := uuid.New()
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now().Add(-time.Minute)

	tests := []struct {
		name      string
		userID    uuid.UUID
		setup     func(pool *MockPool, row *MockRow)
		wantIs    error
		wantNotIs error
		wantUser  domain.User
	}{
		{
			name:   "success scans user by id",
			userID: id,
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
						[]any{id},
					).
					Return(row).
					Once()

				row.EXPECT().
					Scan(mock.Anything).
					Run(func(dest ...any) {
						ptrs, ok := dest[0].([]any)
						if !ok {
							return
						}

						idPtr, ok := ptrs[0].(*uuid.UUID)
						if !ok {
							return
						}

						usernamePtr, ok := ptrs[1].(*string)
						if !ok {
							return
						}

						passwordHashPtr, ok := ptrs[2].(*string)
						if !ok {
							return
						}

						rolePtr, ok := ptrs[3].(*string)
						if !ok {
							return
						}

						createdAtPtr, ok := ptrs[4].(*time.Time)
						if !ok {
							return
						}

						updatedAtPtr, ok := ptrs[5].(**time.Time)
						if !ok {
							return
						}

						*idPtr = id
						*usernamePtr = "alice"
						*passwordHashPtr = "hash"
						*rolePtr = "admin"
						*createdAtPtr = createdAt
						*updatedAtPtr = &updatedAt
					}).
					Return(nil).
					Once()
			},
			wantUser: mustNewUser(t,
				id,
				"alice",
				"hash",
				"admin",
				createdAt,
				&updatedAt,
			),
		},
		{
			name:   "success scans user with nil updated_at",
			userID: id,
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
					Run(func(dest ...any) {
						ptrs, ok := dest[0].([]any)
						if !ok {
							return
						}

						idPtr, ok := ptrs[0].(*uuid.UUID)
						if !ok {
							return
						}

						usernamePtr, ok := ptrs[1].(*string)
						if !ok {
							return
						}

						passwordHashPtr, ok := ptrs[2].(*string)
						if !ok {
							return
						}

						rolePtr, ok := ptrs[3].(*string)
						if !ok {
							return
						}

						createdAtPtr, ok := ptrs[4].(*time.Time)
						if !ok {
							return
						}

						*idPtr = id
						*usernamePtr = "bob"
						*passwordHashPtr = "hash"
						*rolePtr = "user"
						*createdAtPtr = createdAt
					}).
					Return(nil).
					Once()
			},
			wantUser: mustNewUser(t,
				id,
				"bob",
				"hash",
				"user",
				createdAt,
				nil,
			),
		},
		{
			name:   "no rows maps to ErrUserNotFound",
			userID: id,
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
					Return(core_postgres_pool.ErrNoRows).
					Once()
			},
			wantIs:    errs.ErrUserNotFound,
			wantNotIs: core_postgres_pool.ErrNoRows,
		},
		{
			name:   "scan error is wrapped",
			userID: id,
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

			user, err := repo.GetUserByID(context.Background(), tt.userID)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				if tt.wantNotIs != nil {
					must.NotErrorIs(err, tt.wantNotIs)
				}

				return
			}

			must.NoError(err)
			is.Equal(tt.wantUser, user)
		})
	}
}
