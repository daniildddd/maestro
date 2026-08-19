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

func newTestServiceWithHasher(
	usersRepository service.UsersRepository,
	passwordHasher service.PasswordHasher,
) *service.UsersService {
	return service.NewUsersService(usersRepository, passwordHasher)
}

func TestDeleteMe(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	createdAt := time.Now().Add(-time.Hour)
	password := "secret123"

	tests := []struct {
		name       string
		setupMocks func(repo *MockUsersRepository, hasher *MockPasswordHasher)
		wantIs     error
	}{
		{
			name: "success verifies password and deletes account",
			setupMocks: func(repo *MockUsersRepository, hasher *MockPasswordHasher) {
				repo.EXPECT().
					GetUserByID(mock.Anything, userID).
					Return(mustNewUser(t,
						userID,
						"alice",
						"hash",
						"user",
						createdAt,
						nil,
					), nil).
					Once()

				hasher.EXPECT().
					Verify("hash", password).
					Return(nil).
					Once()

				repo.EXPECT().
					DeleteUser(mock.Anything, userID).
					Return(nil).
					Once()
			},
		},
		{
			name: "wrong password returns ErrInvalidCredentials",
			setupMocks: func(repo *MockUsersRepository, hasher *MockPasswordHasher) {
				repo.EXPECT().
					GetUserByID(mock.Anything, userID).
					Return(mustNewUser(t,
						userID,
						"alice",
						"hash",
						"user",
						createdAt,
						nil,
					), nil).
					Once()

				hasher.EXPECT().
					Verify("hash", password).
					Return(errs.ErrInternal).
					Once()
			},
			wantIs: errs.ErrInvalidCredentials,
		},
		{
			name: "user not found is passed through",
			setupMocks: func(repo *MockUsersRepository, _ *MockPasswordHasher) {
				repo.EXPECT().
					GetUserByID(mock.Anything, userID).
					Return(domain.User{}, errs.ErrUserNotFound).
					Once()
			},
			wantIs: errs.ErrUserNotFound,
		},
		{
			name: "delete error is wrapped",
			setupMocks: func(repo *MockUsersRepository, hasher *MockPasswordHasher) {
				repo.EXPECT().
					GetUserByID(mock.Anything, userID).
					Return(mustNewUser(t,
						userID,
						"alice",
						"hash",
						"user",
						createdAt,
						nil,
					), nil).
					Once()

				hasher.EXPECT().
					Verify("hash", password).
					Return(nil).
					Once()

				repo.EXPECT().
					DeleteUser(mock.Anything, userID).
					Return(errs.ErrInternal).
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
			tt.setupMocks(repo, hasher)

			svc := newTestServiceWithHasher(repo, hasher)

			err := svc.DeleteMe(context.Background(), userID, password)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
		})
	}
}
