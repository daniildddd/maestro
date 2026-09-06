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

type userRowFixture struct {
	id           uuid.UUID
	username     string
	passwordHash string
	role         string
	createdAt    time.Time
	updatedAt    *time.Time
}

//nolint:unparam // test helper mirrors the users table row; any argument may vary per test
func mustUserRow(
	id uuid.UUID,
	username string,
	passwordHash string,
	role string,
	createdAt time.Time,
	updatedAt *time.Time,
) userRowFixture {
	return userRowFixture{
		id:           id,
		username:     username,
		passwordHash: passwordHash,
		role:         role,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

//nolint:unparam // test helper mirrors the users table row; any argument may vary per test
func scanUserIntoRow(row *MockRow, fixture userRowFixture) {
	row.EXPECT().
		Scan(mock.Anything).
		Run(func(dest ...any) {
			fillUserDest(dest, fixture)
		}).
		Return(nil).
		Once()
}

func scanUserPairIntoRow(row *MockRow, before, after userRowFixture) {
	row.EXPECT().
		Scan(mock.Anything).
		Run(func(dest ...any) {
			fillUserDest(dest[:6], before)
			fillUserDest(dest[6:], after)
		}).
		Return(nil).
		Once()
}

func scanUserIntoRows(rows *MockRows, fixture userRowFixture) {
	rows.EXPECT().
		Scan(mock.Anything).
		Run(func(dest ...any) {
			fillUserDest(dest, fixture)
		}).
		Return(nil).
		Once()
}

func fillUserDest(ptrs []any, row userRowFixture) {
	if idPtr, ok := ptrs[0].(*uuid.UUID); ok {
		*idPtr = row.id
	}

	if usernamePtr, ok := ptrs[1].(*string); ok {
		*usernamePtr = row.username
	}

	if passwordHashPtr, ok := ptrs[2].(*string); ok {
		*passwordHashPtr = row.passwordHash
	}

	if rolePtr, ok := ptrs[3].(*string); ok {
		*rolePtr = row.role
	}

	if createdAtPtr, ok := ptrs[4].(*time.Time); ok {
		*createdAtPtr = row.createdAt
	}

	if updatedAtPtr, ok := ptrs[5].(**time.Time); ok {
		*updatedAtPtr = row.updatedAt
	}
}
