package domain_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func TestNewUser(t *testing.T) {
	t.Parallel()

	createdAt := time.Now()
	updatedAt := createdAt.Add(time.Minute)

	tests := []struct {
		name         string
		username     string
		passwordHash string
		role         string
		wantErr      error
	}{
		{
			name:         "valid inputs return user",
			username:     "alice",
			passwordHash: "hashed-password-123",
			role:         "admin",
			wantErr:      nil,
		},
		{
			name:         "username too short rejected",
			username:     "ab",
			passwordHash: "hashed-password-123",
			role:         "admin",
			wantErr:      domain.ErrInvalidUsername,
		},
		{
			name:         "unknown role rejected",
			username:     "alice",
			passwordHash: "hashed-password-123",
			role:         "root",
			wantErr:      domain.ErrInvalidRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			must := require.New(t)

			got, err := domain.NewUser(
				uuid.New(),
				tt.username,
				tt.passwordHash,
				tt.role,
				createdAt,
				&updatedAt,
			)

			if tt.wantErr != nil {
				must.ErrorIs(err, tt.wantErr)
				is.Equal(domain.User{}, got)

				return
			}

			must.NoError(err)

			is.Equal(tt.username, got.Username)
			is.Equal(tt.passwordHash, got.PasswordHash)
			is.Equal(tt.role, got.Role)
			is.True(got.CreatedAt.Equal(createdAt))
			must.NotNil(got.UpdatedAt)
			is.True(got.UpdatedAt.Equal(updatedAt))
		})
	}
}

func TestUser_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		username string
		role     string
		wantErr  error
	}{
		{
			name:     "valid user returns no error",
			username: "alice",
			role:     "admin",
			wantErr:  nil,
		},
		{
			name:     "valid user with user role returns no error",
			username: "alice",
			role:     "user",
			wantErr:  nil,
		},
		{
			name:     "username too short rejected",
			username: "ab",
			role:     "admin",
			wantErr:  domain.ErrInvalidUsername,
		},
		{
			name:     "username too long rejected",
			username: strings.Repeat("a", 33),
			role:     "admin",
			wantErr:  domain.ErrInvalidUsername,
		},
		{
			name:     "username with invalid characters rejected",
			username: "bad name!",
			role:     "admin",
			wantErr:  domain.ErrInvalidUsername,
		},
		{
			name:     "unknown role rejected",
			username: "alice",
			role:     "root",
			wantErr:  domain.ErrInvalidRole,
		},
		{
			name:     "empty role rejected",
			username: "alice",
			role:     "",
			wantErr:  domain.ErrInvalidRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			user := domain.User{
				Username: tt.username,
				Role:     tt.role,
			}

			err := user.Validate()

			if tt.wantErr != nil {
				must.ErrorIs(err, tt.wantErr)

				return
			}

			must.NoError(err)
		})
	}
}
