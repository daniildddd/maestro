package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/features/users/repository"
)

const opTimeoutToDomain = 100 * time.Millisecond

func TestUsersRepository_CreateUserToDomainError(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	pool := NewMockPool(t)
	row := NewMockRow(t)

	pool.EXPECT().
		OpTimeout().
		Return(opTimeoutToDomain).
		Once()

	pool.EXPECT().
		QueryRow(mock.Anything, mock.Anything, mock.Anything).
		Return(row).
		Once()

	scanUserIntoRow(row, mustUserRow(uuid.New(), "alice", "hash", "root", time.Time{}, nil))

	repo := repository.NewUsersRepository(pool)

	_, err := repo.CreateUser(context.Background(), domain.User{
		ID:           uuid.New(),
		Username:     "alice",
		PasswordHash: "hash",
		Role:         "root",
	})
	must.Error(err)
	must.ErrorContains(err, "map user")
	must.ErrorIs(err, domain.ErrInvalidRole)
}

func TestUsersRepository_DeleteUserToDomainError(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	pool := NewMockPool(t)
	row := NewMockRow(t)

	pool.EXPECT().
		OpTimeout().
		Return(opTimeoutToDomain).
		Once()

	pool.EXPECT().
		QueryRow(mock.Anything, mock.Anything, mock.Anything).
		Return(row).
		Once()

	scanUserIntoRow(row, mustUserRow(uuid.New(), "alice", "hash", "root", time.Time{}, nil))

	repo := repository.NewUsersRepository(pool)

	_, err := repo.DeleteUser(context.Background(), uuid.New())
	must.Error(err)
	must.ErrorContains(err, "map user")
	must.ErrorIs(err, domain.ErrInvalidRole)
}

func TestUsersRepository_GetUserByIDToDomainError(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	pool := NewMockPool(t)
	row := NewMockRow(t)

	pool.EXPECT().
		OpTimeout().
		Return(opTimeoutToDomain).
		Once()

	pool.EXPECT().
		QueryRow(mock.Anything, mock.Anything, mock.Anything).
		Return(row).
		Once()

	scanUserIntoRow(row, mustUserRow(uuid.New(), "alice", "hash", "root", time.Time{}, nil))

	repo := repository.NewUsersRepository(pool)

	_, err := repo.GetUserByID(context.Background(), uuid.New())
	must.Error(err)
	must.ErrorContains(err, "map user")
	must.ErrorIs(err, domain.ErrInvalidRole)
}

func TestUsersRepository_UpdateUserToDomainError(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	pool := NewMockPool(t)
	row := NewMockRow(t)

	pool.EXPECT().
		OpTimeout().
		Return(opTimeoutToDomain).
		Once()

	pool.EXPECT().
		QueryRow(mock.Anything, mock.Anything, mock.Anything).
		Return(row).
		Once()

	row.EXPECT().
		Scan(mock.Anything).
		Run(func(dest ...any) {
			fillUserDest(dest[:6], mustUserRow(uuid.New(), "alice", "hash", "root", time.Time{}, nil))
			fillUserDest(dest[6:], mustUserRow(uuid.New(), "alice", "hash", "root", time.Time{}, nil))
		}).
		Return(nil).
		Once()

	repo := repository.NewUsersRepository(pool)

	_, _, err := repo.UpdateUser(context.Background(), uuid.New(), "newusername")
	must.Error(err)
	must.ErrorContains(err, "map user")
	must.ErrorIs(err, domain.ErrInvalidRole)
}

func TestUsersRepository_UpdateUserAfterToDomainError(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	pool := NewMockPool(t)
	row := NewMockRow(t)

	pool.EXPECT().
		OpTimeout().
		Return(opTimeoutToDomain).
		Once()

	pool.EXPECT().
		QueryRow(mock.Anything, mock.Anything, mock.Anything).
		Return(row).
		Once()

	valid := mustUserRow(uuid.New(), "validuser", "hash", "admin", time.Now().Add(-time.Hour), nil)

	row.EXPECT().
		Scan(mock.Anything).
		Run(func(dest ...any) {
			fillUserDest(dest[:6], valid)
			fillUserDest(dest[6:], mustUserRow(uuid.New(), "alice", "hash", "root", time.Time{}, nil))
		}).
		Return(nil).
		Once()

	repo := repository.NewUsersRepository(pool)

	_, _, err := repo.UpdateUser(context.Background(), uuid.New(), "newusername")
	must.Error(err)
	must.ErrorContains(err, "map user")
	must.ErrorIs(err, domain.ErrInvalidRole)
}

func TestUsersRepository_GetUsersToDomainError(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	pool := NewMockPool(t)
	rows := NewMockRows(t)

	pool.EXPECT().
		OpTimeout().
		Return(opTimeoutToDomain).
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

	repo := repository.NewUsersRepository(pool)

	_, err := repo.GetUsers(context.Background(), &domain.UserFilter{Page: 1, Limit: 20})
	must.Error(err)
	must.ErrorContains(err, "map user")
	must.ErrorIs(err, domain.ErrInvalidRole)
}
