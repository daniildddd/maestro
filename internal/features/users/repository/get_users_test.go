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
	"github.com/daniildddd/maestro/internal/features/users/repository"
)

//nolint:maintidx // table-driven test: complexity comes from mock setup Run blocks
func TestGetUsers(t *testing.T) {
	t.Parallel()

	const opTimeout = 100 * time.Millisecond

	userID := uuid.New()
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now().Add(-time.Minute)

	tests := []struct {
		name      string
		filter    domain.UserFilter
		setup     func(pool *MockPool, rows *MockRows)
		wantIs    error
		wantUsers []domain.User
	}{
		{
			name:   "success returns users",
			filter: domain.UserFilter{Page: 1, Limit: 20},
			setup: func(pool *MockPool, rows *MockRows) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Query(
						mock.MatchedBy(func(ctx context.Context) bool {
							deadline, ok := ctx.Deadline()
							if !ok {
								return false
							}

							return time.Until(deadline) <= opTimeout
						}),
						"\n\tSELECT id, username, password_hash, role, created_at, updated_at\n\tFROM users ORDER BY created_at LIMIT $1 OFFSET $2",
						[]any{20, 0},
					).
					Return(rows, nil).
					Once()

				rows.EXPECT().
					Next().
					Return(true).
					Once()

				rows.EXPECT().
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

						*idPtr = userID
						*usernamePtr = "alice"
						*passwordHashPtr = "hash"
						*rolePtr = "admin"
						*createdAtPtr = createdAt
						*updatedAtPtr = &updatedAt
					}).
					Return(nil).
					Once()

				rows.EXPECT().
					Next().
					Return(false).
					Once()

				rows.EXPECT().
					Err().
					Return(nil).
					Once()

				rows.EXPECT().
					Close().
					Once()
			},
			wantUsers: []domain.User{
				mustNewUser(t,
					userID,
					"alice",
					"hash",
					"admin",
					createdAt,
					&updatedAt,
				),
			},
		},
		{
			name:   "success with username and role filters",
			filter: domain.UserFilter{Page: 2, Limit: 10, Username: "alice", UserRole: "admin"},
			setup: func(pool *MockPool, rows *MockRows) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Query(
						mock.Anything,
						"\n\tSELECT id, username, password_hash, role, created_at, updated_at\n\tFROM users WHERE username=$1 AND role=$2 ORDER BY created_at LIMIT $3 OFFSET $4",
						[]any{"alice", "admin", 10, 10},
					).
					Return(rows, nil).
					Once()

				rows.EXPECT().
					Next().
					Return(false).
					Once()

				rows.EXPECT().
					Err().
					Return(nil).
					Once()

				rows.EXPECT().
					Close().
					Once()
			},
			wantUsers: nil,
		},
		{
			name:   "success with username filter only",
			filter: domain.UserFilter{Page: 1, Limit: 20, Username: "alice"},
			setup: func(pool *MockPool, rows *MockRows) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Query(
						mock.Anything,
						"\n\tSELECT id, username, password_hash, role, created_at, updated_at\n\tFROM users WHERE username=$1 ORDER BY created_at LIMIT $2 OFFSET $3",
						[]any{"alice", 20, 0},
					).
					Return(rows, nil).
					Once()

				rows.EXPECT().
					Next().
					Return(false).
					Once()

				rows.EXPECT().
					Err().
					Return(nil).
					Once()

				rows.EXPECT().
					Close().
					Once()
			},
			wantUsers: nil,
		},
		{
			name:   "success with role filter only",
			filter: domain.UserFilter{Page: 1, Limit: 20, UserRole: "admin"},
			setup: func(pool *MockPool, rows *MockRows) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Query(
						mock.Anything,
						"\n\tSELECT id, username, password_hash, role, created_at, updated_at\n\tFROM users WHERE role=$1 ORDER BY created_at LIMIT $2 OFFSET $3",
						[]any{"admin", 20, 0},
					).
					Return(rows, nil).
					Once()

				rows.EXPECT().
					Next().
					Return(false).
					Once()

				rows.EXPECT().
					Err().
					Return(nil).
					Once()

				rows.EXPECT().
					Close().
					Once()
			},
			wantUsers: nil,
		},
		{
			name:   "query error is wrapped",
			filter: domain.UserFilter{Page: 1, Limit: 20},
			setup: func(pool *MockPool, _ *MockRows) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Query(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errs.ErrInternal).
					Once()
			},
			wantIs: errs.ErrInternal,
		},
		{
			name:   "scan error is wrapped",
			filter: domain.UserFilter{Page: 1, Limit: 20},
			setup: func(pool *MockPool, rows *MockRows) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Query(mock.Anything, mock.Anything, mock.Anything).
					Return(rows, nil).
					Once()

				rows.EXPECT().
					Next().
					Return(true).
					Once()

				rows.EXPECT().
					Scan(mock.Anything).
					Return(errs.ErrInternal).
					Once()

				rows.EXPECT().
					Close().
					Once()
			},
			wantIs: errs.ErrInternal,
		},
		{
			name:   "rows error is wrapped",
			filter: domain.UserFilter{Page: 1, Limit: 20},
			setup: func(pool *MockPool, rows *MockRows) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Query(mock.Anything, mock.Anything, mock.Anything).
					Return(rows, nil).
					Once()

				rows.EXPECT().
					Next().
					Return(false).
					Once()

				rows.EXPECT().
					Err().
					Return(errs.ErrInternal).
					Once()

				rows.EXPECT().
					Close().
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
			rows := NewMockRows(t)
			tt.setup(pool, rows)

			repo := repository.NewUsersRepository(pool)

			users, err := repo.GetUsers(context.Background(), tt.filter)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			is.Equal(tt.wantUsers, users)
		})
	}
}
