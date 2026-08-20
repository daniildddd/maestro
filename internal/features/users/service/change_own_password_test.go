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

func TestChangeOwnPassword(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	createdAt := time.Now().Add(-time.Hour)

	tests := []struct {
		name       string
		oldPass    string
		newPass    string
		setupMocks func(repo *MockUsersRepository, hasher *MockPasswordHasher)
		wantIs     error
	}{
		{
			name:    "success verifies old and updates new password",
			oldPass: "oldSecret123",
			newPass: "newSecret123",
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
					Verify("hash", "oldSecret123").
					Return(nil).
					Once()

				hasher.EXPECT().
					Hash("newSecret123").
					Return("new-hash", nil).
					Once()

				repo.EXPECT().
					ChangePassword(mock.Anything, userID, "new-hash").
					Return(nil).
					Once()
			},
		},
		{
			name:    "invalid new password returns ErrValidationFailed",
			oldPass: "oldSecret123",
			newPass: "short",
			setupMocks: func(_ *MockUsersRepository, _ *MockPasswordHasher) {
			},
			wantIs: errs.ErrValidationFailed,
		},
		{
			name:    "user not found is passed through",
			oldPass: "oldSecret123",
			newPass: "newSecret123",
			setupMocks: func(repo *MockUsersRepository, _ *MockPasswordHasher) {
				repo.EXPECT().
					GetUserByID(mock.Anything, userID).
					Return(domain.User{}, errs.ErrUserNotFound).
					Once()
			},
			wantIs: errs.ErrUserNotFound,
		},
		{
			name:    "wrong old password returns ErrInvalidCredentials",
			oldPass: "wrong-pass",
			newPass: "newSecret123",
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
					Verify("hash", "wrong-pass").
					Return(errs.ErrInternal).
					Once()
			},
			wantIs: errs.ErrInvalidCredentials,
		},
		{
			name:    "hash error is wrapped",
			oldPass: "oldSecret123",
			newPass: "newSecret123",
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
					Verify("hash", "oldSecret123").
					Return(nil).
					Once()

				hasher.EXPECT().
					Hash("newSecret123").
					Return("", errs.ErrInternal).
					Once()
			},
			wantIs: errs.ErrInternal,
		},
		{
			name:    "repository error is wrapped",
			oldPass: "oldSecret123",
			newPass: "newSecret123",
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
					Verify("hash", "oldSecret123").
					Return(nil).
					Once()

				hasher.EXPECT().
					Hash("newSecret123").
					Return("new-hash", nil).
					Once()

				repo.EXPECT().
					ChangePassword(mock.Anything, userID, "new-hash").
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

			err := svc.ChangeOwnPassword(context.Background(), userID, tt.oldPass, tt.newPass)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
		})
	}
}
