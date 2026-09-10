package service_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
)

func TestUsersService_DeleteUser(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	createdAt := time.Now().Add(-time.Hour)

	tests := []struct {
		name      string
		setupMock func(repo *MockUsersRepository)
		wantIs    error
	}{
		{
			name: "success deletes user and records audit",
			setupMock: func(repo *MockUsersRepository) {
				repo.EXPECT().
					DeleteUser(mock.Anything, userID).
					Return(mustNewUser(t, userID, "deleteduser", "hash", "user", createdAt, nil), nil).
					Once()
			},
		},
		{
			name: "repository error is wrapped",
			setupMock: func(repo *MockUsersRepository) {
				repo.EXPECT().
					DeleteUser(mock.Anything, userID).
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
			tt.setupMock(repo)

			svc := newTestService(t, repo)

			err := svc.DeleteUser(testCtx(), userID)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
		})
	}
}
