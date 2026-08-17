package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func TestNewUser(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	id := uuid.New()
	createdAt := time.Now()
	updatedAt := createdAt.Add(time.Minute)
	expectedUpdatedAt := updatedAt
	got := domain.NewUser(
		id,
		"alice",
		"hashed-password-123",
		"admin",
		createdAt,
		&updatedAt,
	)

	is.Equal(id, got.ID)
	is.Equal("alice", got.Username)
	is.Equal("hashed-password-123", got.PasswordHash)
	is.Equal("admin", got.Role)
	is.True(got.CreatedAt.Equal(createdAt))
	must.NotNil(got.UpdatedAt)
	is.True(got.UpdatedAt.Equal(expectedUpdatedAt))
}
