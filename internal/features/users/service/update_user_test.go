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
)

func TestUpdateUser(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	createdAt := time.Now().Add(-time.Hour)

	tests := []struct {
		name       string
		username   string
		setupMocks func(repo *MockUsersRepository, hasher *MockPasswordHasher)
		wantIs     error
		wantUser   domain.User
	}{
		{
			name:     "success updates username",
			username: "updateduser",
			setupMocks: func(repo *MockUsersRepository, _ *MockPasswordHasher) {
				repo.EXPECT().
					UpdateUser(mock.Anything, userID, "updateduser").
					Return(mustNewUser(t,
						userID,
						"updateduser",
						"hash",
						"user",
						createdAt,
						nil,
					), nil).
					Once()
			},
			wantUser: mustNewUser(t, userID, "updateduser", "hash", "user", createdAt, nil),
		},
		{
			name:     "invalid username returns ErrValidationFailed",
			username: "ab",
			setupMocks: func(_ *MockUsersRepository, _ *MockPasswordHasher) {
			},
			wantIs: errs.ErrValidationFailed,
		},
		{
			name:     "repository error is wrapped",
			username: "updateduser",
			setupMocks: func(repo *MockUsersRepository, _ *MockPasswordHasher) {
				repo.EXPECT().
					UpdateUser(mock.Anything, userID, "updateduser").
					Return(domain.User{}, errs.ErrUserNotFound).
					Once()
			},
			wantIs: errs.ErrUserNotFound,
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

			user, err := svc.UpdateUser(context.Background(), userID, tt.username)

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
