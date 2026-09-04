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
		{
			name:         "username too long rejected",
			username:     strings.Repeat("a", 33),
			passwordHash: "hashed-password-123",
			role:         "admin",
			wantErr:      domain.ErrInvalidUsername,
		},
		{
			name:         "username with invalid characters rejected",
			username:     "bad name!",
			passwordHash: "hashed-password-123",
			role:         "admin",
			wantErr:      domain.ErrInvalidUsername,
		},
		{
			name:         "username with non-ascii rejected",
			username:     "пользователь",
			passwordHash: "hashed-password-123",
			role:         "admin",
			wantErr:      domain.ErrInvalidUsername,
		},
		{
			name:         "username with emoji rejected",
			username:     "user😀",
			passwordHash: "hashed-password-123",
			role:         "admin",
			wantErr:      domain.ErrInvalidUsername,
		},
		{
			name:         "empty role rejected",
			username:     "alice",
			passwordHash: "hashed-password-123",
			role:         "",
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

func TestNewUserFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		page     int
		limit    int
		username string
		userRole string
		wantPage int
		wantLim  int
		wantErr  error
	}{
		{
			name:     "valid username filter accepted",
			page:     1,
			limit:    20,
			username: "user_name-1",
			wantPage: 1,
			wantLim:  20,
		},
		{
			name:     "defaults applied",
			page:     0,
			limit:    0,
			wantPage: 1,
			wantLim:  20,
		},
		{
			name:     "values kept",
			page:     3,
			limit:    50,
			username: "alice",
			userRole: "admin",
			wantPage: 3,
			wantLim:  50,
		},
		{
			name:     "negative page normalized",
			page:     -5,
			limit:    10,
			wantPage: 1,
			wantLim:  10,
		},
		{
			name:     "negative limit normalized",
			page:     1,
			limit:    -5,
			wantPage: 1,
			wantLim:  20,
		},
		{
			name:     "limit capped at 100",
			page:     1,
			limit:    500,
			wantPage: 1,
			wantLim:  100,
		},
		{
			name:     "unknown role rejected",
			page:     1,
			limit:    20,
			userRole: "root",
			wantErr:  domain.ErrInvalidRole,
		},
		{
			name:     "empty role accepted",
			page:     1,
			limit:    20,
			userRole: "",
			wantPage: 1,
			wantLim:  20,
		},
		{
			name:     "invalid username rejected",
			page:     1,
			limit:    20,
			username: "bad name!",
			wantErr:  domain.ErrInvalidUsername,
		},
		{
			name:     "non-ascii username rejected",
			page:     1,
			limit:    20,
			username: "пользователь",
			wantErr:  domain.ErrInvalidUsername,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			filter, err := domain.NewUserFilter(tt.page, tt.limit, tt.username, tt.userRole)

			if tt.wantErr != nil {
				must.ErrorIs(err, tt.wantErr)
				must.Nil(filter)

				return
			}

			must.NoError(err)
			must.Equal(tt.wantPage, filter.Page)
			must.Equal(tt.wantLim, filter.Limit)
			must.Equal(tt.username, filter.Username)
			must.Equal(tt.userRole, filter.UserRole)
		})
	}
}

func TestValidatePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "valid password",
			password: "secret123",
			wantErr:  nil,
		},
		{
			name:     "valid password with spaces",
			password: "my pass phrase",
			wantErr:  nil,
		},
		//nolint:gosec // test fixture: unicode password string, not a real credential
		{
			name:     "valid password with unicode",
			password: "секретный пароль",
			wantErr:  nil,
		},
		{
			name:     "valid password exactly 8 chars",
			password: "12345678",
			wantErr:  nil,
		},
		{
			name:     "valid password exactly 128 chars",
			password: strings.Repeat("a", 128),
			wantErr:  nil,
		},
		{
			name:     "password too short rejected",
			password: "1234567",
			wantErr:  domain.ErrInvalidPassword,
		},
		{
			name:     "password too long rejected",
			password: strings.Repeat("a", 129),
			wantErr:  domain.ErrInvalidPassword,
		},
		{
			name:     "empty password rejected",
			password: "",
			wantErr:  domain.ErrInvalidPassword,
		},
		{
			name:     "password with newline rejected",
			password: "secret\n123",
			wantErr:  domain.ErrInvalidPassword,
		},
		{
			name:     "password with tab rejected",
			password: "secret\t123",
			wantErr:  domain.ErrInvalidPassword,
		},
		{
			name:     "password with del char rejected",
			password: "secret\x7F123",
			wantErr:  domain.ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			err := domain.ValidatePassword(tt.password)

			if tt.wantErr != nil {
				must.ErrorIs(err, tt.wantErr)

				return
			}

			must.NoError(err)
		})
	}
}

func TestValidateUsername(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		username string
		wantErr  error
	}{
		{
			name:     "valid username",
			username: "alice",
			wantErr:  nil,
		},
		{
			name:     "valid username with digits and separators",
			username: "user_name-1",
			wantErr:  nil,
		},
		{
			name:     "minimum length accepted",
			username: "abc",
			wantErr:  nil,
		},
		{
			name:     "maximum length accepted",
			username: strings.Repeat("a", 32),
			wantErr:  nil,
		},
		{
			name:     "empty username rejected",
			username: "",
			wantErr:  domain.ErrInvalidUsername,
		},
		{
			name:     "username too short rejected",
			username: "ab",
			wantErr:  domain.ErrInvalidUsername,
		},
		{
			name:     "username too long rejected",
			username: strings.Repeat("a", 33),
			wantErr:  domain.ErrInvalidUsername,
		},
		{
			name:     "username with space rejected",
			username: "bad name",
			wantErr:  domain.ErrInvalidUsername,
		},
		{
			name:     "username with dot rejected",
			username: "user.name",
			wantErr:  domain.ErrInvalidUsername,
		},
		{
			name:     "username with non-ascii rejected",
			username: "пользователь",
			wantErr:  domain.ErrInvalidUsername,
		},
		{
			name:     "username with at sign rejected",
			username: "user@mail",
			wantErr:  domain.ErrInvalidUsername,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			err := domain.ValidateUsername(tt.username)

			if tt.wantErr != nil {
				must.ErrorIs(err, tt.wantErr)

				return
			}

			must.NoError(err)
		})
	}
}
