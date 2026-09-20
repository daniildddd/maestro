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

func TestUsersRepository_GetUsersByIDs(t *testing.T) {
	t.Parallel()

	const opTimeout = 100 * time.Millisecond

	aliceID := uuid.New()
	bobID := uuid.New()
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now().Add(-time.Minute)

	tests := []struct {
		name      string
		ids       []uuid.UUID
		setup     func(pool *MockPool, rows *MockRows)
		wantIs    error
		wantUsers []domain.User
	}{
		{
			name: "success returns users",
			ids:  []uuid.UUID{aliceID, bobID},
			setup: func(pool *MockPool, rows *MockRows) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Query(
						mock.Anything,
						"\n\tSELECT id, username, password_hash, role, created_at, updated_at\n"+
							"\tFROM users WHERE id = ANY($1)",
						[]any{[]uuid.UUID{aliceID, bobID}},
					).
					Return(rows, nil).
					Once()

				rows.EXPECT().
					Next().
					Return(true).
					Once()

				scanUserIntoRows(rows, mustUserRow(aliceID, "alice", "hash", "admin", createdAt, &updatedAt))

				rows.EXPECT().
					Next().
					Return(true).
					Once()

				scanUserIntoRows(rows, mustUserRow(bobID, "bob", "hash", "user", createdAt, nil))

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
					aliceID,
					"alice",
					"hash",
					"admin",
					createdAt,
					&updatedAt,
				),
				mustNewUser(t,
					bobID,
					"bob",
					"hash",
					"user",
					createdAt,
					nil,
				),
			},
		},
		{
			name:      "empty ids return nil without query",
			ids:       nil,
			setup:     func(_ *MockPool, _ *MockRows) {},
			wantUsers: nil,
		},
		{
			name: "query error is wrapped",
			ids:  []uuid.UUID{aliceID},
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
			name: "scan error is wrapped",
			ids:  []uuid.UUID{aliceID},
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
			name: "invalid row data maps to validation error",
			ids:  []uuid.UUID{aliceID},
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

				scanUserIntoRows(rows, mustUserRow(uuid.New(), "alice", "hash", "root", time.Time{}, nil))

				rows.EXPECT().
					Close().
					Once()
			},
			wantIs: domain.ErrInvalidRole,
		},
		{
			name: "rows error is wrapped",
			ids:  []uuid.UUID{aliceID},
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

			users, err := repo.GetUsersByIDs(context.Background(), tt.ids)

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
