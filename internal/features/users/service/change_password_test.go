package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/errs"
)

func TestChangePassword(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name       string
		newPass    string
		setupMocks func(repo *MockUsersRepository, hasher *MockPasswordHasher)
		wantIs     error
	}{
		{
			name:    "success hashes and updates password",
			newPass: "newSecret123",
			setupMocks: func(repo *MockUsersRepository, hasher *MockPasswordHasher) {
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
			name:    "invalid password returns ErrValidationFailed",
			newPass: "short",
			setupMocks: func(_ *MockUsersRepository, _ *MockPasswordHasher) {
			},
			wantIs: errs.ErrValidationFailed,
		},
		{
			name:    "hash error is wrapped",
			newPass: "newSecret123",
			setupMocks: func(_ *MockUsersRepository, hasher *MockPasswordHasher) {
				hasher.EXPECT().
					Hash("newSecret123").
					Return("", errs.ErrInternal).
					Once()
			},
			wantIs: errs.ErrInternal,
		},
		{
			name:    "repository error is wrapped",
			newPass: "newSecret123",
			setupMocks: func(repo *MockUsersRepository, hasher *MockPasswordHasher) {
				hasher.EXPECT().
					Hash("newSecret123").
					Return("new-hash", nil).
					Once()

				repo.EXPECT().
					ChangePassword(mock.Anything, userID, "new-hash").
					Return(errs.ErrUserNotFound).
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

			err := svc.ChangePassword(context.Background(), userID, tt.newPass)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
		})
	}
}
