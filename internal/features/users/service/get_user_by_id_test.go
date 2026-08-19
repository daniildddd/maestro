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

func TestGetUserByID(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	createdAt := time.Now().Add(-time.Hour)

	tests := []struct {
		name      string
		setupMock func(repo *MockUsersRepository)
		wantIs    error
		wantUser  domain.User
	}{
		{
			name: "success returns user",
			setupMock: func(repo *MockUsersRepository) {
				repo.EXPECT().
					GetUserByID(mock.Anything, userID).
					Return(mustNewUser(t,
						userID,
						"alice",
						"hash",
						"admin",
						createdAt,
						nil,
					), nil).
					Once()
			},
			wantUser: mustNewUser(t,
				userID,
				"alice",
				"hash",
				"admin",
				createdAt,
				nil,
			),
		},
		{
			name: "repository error passes through unchanged",
			setupMock: func(repo *MockUsersRepository) {
				repo.EXPECT().
					GetUserByID(mock.Anything, userID).
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

			user, err := svc.GetUserByID(context.Background(), userID)

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
