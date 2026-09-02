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

func newTestService(
	t *testing.T,
	usersRepository service.UsersRepository,
) *service.UsersService {
	t.Helper()

	return service.NewUsersService(
		usersRepository,
		NewMockPasswordHasher(t),
		noopAuditor{},
	)
}

func TestGetUsers(t *testing.T) {
	t.Parallel()

	filter := &domain.UserFilter{Page: 1, Limit: 20}

	userID := uuid.New()
	createdAt := time.Now().Add(-time.Hour)

	tests := []struct {
		name        string
		setupMock   func(repo *MockUsersRepository)
		wantIs      error
		wantUsers   []domain.User
		wantHasMore bool
	}{
		{
			name: "success returns users without has more",
			setupMock: func(repo *MockUsersRepository) {
				repo.EXPECT().
					GetUsers(mock.Anything, filter).
					Return([]domain.User{
						mustNewUser(t,
							userID,
							"alice",
							"hash",
							"admin",
							createdAt,
							nil,
						),
					}, nil).
					Once()
			},
			wantUsers: []domain.User{
				mustNewUser(t,
					userID,
					"alice",
					"hash",
					"admin",
					createdAt,
					nil,
				),
			},
		},
		{
			name: "exactly limit rows has no more",
			setupMock: func(repo *MockUsersRepository) {
				users := make([]domain.User, 0, filter.Limit)
				for range filter.Limit {
					users = append(users, mustNewUser(t, userID, "user", "hash", "user", createdAt, nil))
				}

				repo.EXPECT().
					GetUsers(mock.Anything, filter).
					Return(users, nil).
					Once()
			},
			wantUsers: func() []domain.User {
				users := make([]domain.User, 0, filter.Limit)
				for range filter.Limit {
					users = append(users, mustNewUser(t, userID, "user", "hash", "user", createdAt, nil))
				}

				return users
			}(),
		},
		{
			name: "extra row truncates and sets has more",
			setupMock: func(repo *MockUsersRepository) {
				users := make([]domain.User, 0, filter.Limit+1)
				for range filter.Limit + 1 {
					users = append(users, mustNewUser(t, userID, "user", "hash", "user", createdAt, nil))
				}

				repo.EXPECT().
					GetUsers(mock.Anything, filter).
					Return(users, nil).
					Once()
			},
			wantUsers: func() []domain.User {
				users := make([]domain.User, 0, filter.Limit)
				for range filter.Limit {
					users = append(users, mustNewUser(t, userID, "user", "hash", "user", createdAt, nil))
				}

				return users
			}(),
			wantHasMore: true,
		},
		{
			name: "repository error is wrapped",
			setupMock: func(repo *MockUsersRepository) {
				repo.EXPECT().
					GetUsers(mock.Anything, filter).
					Return(nil, errs.ErrInternal).
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
			tt.setupMock(repo)

			svc := newTestService(t, repo)

			users, hasMore, err := svc.GetUsers(context.Background(), filter)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			is.Equal(tt.wantUsers, users)
			is.Equal(tt.wantHasMore, hasMore)
		})
	}
}
