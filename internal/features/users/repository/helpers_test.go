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

func scanUserPairIntoRow(row *MockRow, before, after domain.User) {
	row.EXPECT().
		Scan(mock.Anything).
		Run(func(dest ...any) {
			fillUserDest(dest[:6], before)
			fillUserDest(dest[6:], after)
		}).
		Return(nil).
		Once()
}

func scanUserIntoRows(rows *MockRows, u domain.User) {
	rows.EXPECT().
		Scan(mock.Anything).
		Run(func(dest ...any) {
			fillUserDest(dest, u)
		}).
		Return(nil).
		Once()
}

func fillUserDest(ptrs []any, u domain.User) {
	if idPtr, ok := ptrs[0].(*uuid.UUID); ok {
		*idPtr = u.ID
	}

	if usernamePtr, ok := ptrs[1].(*string); ok {
		*usernamePtr = u.Username
	}

	if passwordHashPtr, ok := ptrs[2].(*string); ok {
		*passwordHashPtr = u.PasswordHash
	}

	if rolePtr, ok := ptrs[3].(*string); ok {
		*rolePtr = u.Role
	}

	if createdAtPtr, ok := ptrs[4].(*time.Time); ok {
		*createdAtPtr = u.CreatedAt
	}

	if updatedAtPtr, ok := ptrs[5].(**time.Time); ok {
		*updatedAtPtr = u.UpdatedAt
	}
}
