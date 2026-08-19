package repository_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/daniildddd/maestro/internal/core/domain"
)

//nolint:unparam // test helper mirrors domain.NewUser; any argument may vary per test
func mustNewUser(
	t *testing.T,
	id uuid.UUID,
	username string,
	passwordHash string,
	role string,
	createdAt time.Time,
	updatedAt *time.Time,
) domain.User {
	t.Helper()

	user, err := domain.NewUser(id, username, passwordHash, role, createdAt, updatedAt)
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	return user
}

//nolint:unparam // test helper mirrors the users table row; any argument may vary per test
func scanUserIntoRow(row *MockRow, u domain.User) {
	row.EXPECT().
		Scan(mock.Anything).
		Run(func(dest ...any) {
			fillUserDest(dest, u)
		}).
		Return(nil).
		Once()
}

func fillUserDest(dest []any, u domain.User) {
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

	*idPtr = u.ID
	*usernamePtr = u.Username
	*passwordHashPtr = u.PasswordHash
	*rolePtr = u.Role
	*createdAtPtr = u.CreatedAt
	*updatedAtPtr = u.UpdatedAt
}
