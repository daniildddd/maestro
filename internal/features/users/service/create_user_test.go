package service_test

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
	"github.com/daniildddd/maestro/internal/features/users/service"
)

func TestCreateUser(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	createdAt := time.Now().Add(-time.Hour)

	tests := []struct {
		name      string
		username  string
		password  string
		role      string
		setupMock func(repo *MockUsersRepository, hasher *MockPasswordHasher)
		wantIs    error
		wantUser  domain.User
	}{
		{
			name:     "success hashes password and saves user",
			username: "alice",
			password: "secret123",
			role:     "admin",
			setupMock: func(repo *MockUsersRepository, hasher *MockPasswordHasher) {
				hasher.EXPECT().
					Hash("secret123").
					Return("hashed-secret123", nil).
					Once()

				repo.EXPECT().
					CreateUser(
						mock.Anything,
						mock.MatchedBy(func(u domain.User) bool {
							return u.Username == "alice" &&
								u.PasswordHash == "hashed-secret123" &&
								u.Role == "admin" &&
								u.ID != uuid.Nil &&
								u.UpdatedAt == nil &&
								!u.CreatedAt.IsZero()
						}),
					).
					Return(mustNewUser(t,
						userID,
						"alice",
						"hashed-secret123",
						"admin",
						createdAt,
						nil,
					), nil).
					Once()
			},
			wantUser: mustNewUser(t,
				userID,
				"alice",
				"hashed-secret123",
				"admin",
				createdAt,
				nil,
			),
		},
		{
			name:     "invalid password returns validation error",
			username: "alice",
			password: "1234567",
			role:     "admin",
			setupMock: func(_ *MockUsersRepository, _ *MockPasswordHasher) {
			},
			wantIs: errs.ErrValidationFailed,
		},
		{
			name:     "invalid role returns validation error",
			username: "alice",
			password: "secret123",
			role:     "root",
			setupMock: func(_ *MockUsersRepository, hasher *MockPasswordHasher) {
				hasher.EXPECT().
					Hash("secret123").
					Return("hashed-secret123", nil).
					Once()
			},
			wantIs: errs.ErrValidationFailed,
		},
		{
			name:     "hasher error is wrapped",
			username: "alice",
			password: "secret123",
			role:     "admin",
			setupMock: func(_ *MockUsersRepository, hasher *MockPasswordHasher) {
				hasher.EXPECT().
					Hash("secret123").
					Return("", errs.ErrInternal).
					Once()
			},
			wantIs: errs.ErrInternal,
		},
		{
			name:     "duplicate username propagates conflict",
			username: "alice",
			password: "secret123",
			role:     "admin",
			setupMock: func(repo *MockUsersRepository, hasher *MockPasswordHasher) {
				hasher.EXPECT().
					Hash("secret123").
					Return("hashed-secret123", nil).
					Once()

				repo.EXPECT().
					CreateUser(mock.Anything, mock.Anything).
					Return(domain.User{}, errs.ErrUsernameConflict).
					Once()
			},
			wantIs: errs.ErrUsernameConflict,
		},
		{
			name:     "repository error is wrapped",
			username: "alice",
			password: "secret123",
			role:     "admin",
			setupMock: func(repo *MockUsersRepository, hasher *MockPasswordHasher) {
				hasher.EXPECT().
					Hash("secret123").
					Return("hashed-secret123", nil).
					Once()

				repo.EXPECT().
					CreateUser(mock.Anything, mock.Anything).
					Return(domain.User{}, errs.ErrInternal).
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

			repo := NewMockUsersRepository(t)
			hasher := NewMockPasswordHasher(t)
			tt.setupMock(repo, hasher)

			svc := service.NewUsersService(repo, hasher)

			user, err := svc.CreateUser(
				context.Background(),
				tt.username,
				tt.password,
				tt.role,
			)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			is.Equal(tt.wantUser, user)
		})
	}
}
