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

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name      string
		setupMock func(repo *MockUsersRepository)
		wantIs    error
	}{
		{
			name: "success deletes user",
			setupMock: func(repo *MockUsersRepository) {
				repo.EXPECT().
					DeleteUser(mock.Anything, userID).
					Return(nil).
					Once()
			},
		},
		{
			name: "repository error is wrapped",
			setupMock: func(repo *MockUsersRepository) {
				repo.EXPECT().
					DeleteUser(mock.Anything, userID).
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
			tt.setupMock(repo)

			svc := newTestService(t, repo)

			err := svc.DeleteUser(context.Background(), userID)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
		})
	}
}
