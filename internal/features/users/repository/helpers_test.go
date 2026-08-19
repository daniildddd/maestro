package repository_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

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
