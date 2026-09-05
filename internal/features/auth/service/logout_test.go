package service_test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
)

func TestLogout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rawToken   string
		setupMocks func(
			repo *MockAuthRepository,
			refreshGen *MockRefreshTokenManager,
		)
		wantErrIs error
	}{
		{
			name:     "success deletes token by hash and records audit",
			rawToken: "raw-token",
			setupMocks: func(
				repo *MockAuthRepository,
				refreshGen *MockRefreshTokenManager,
			) {
				refreshGen.EXPECT().
					Hash("raw-token").
					Return("hashed-token").
					Once()

				repo.EXPECT().
					GetRefreshTokenByHash(mock.Anything, "hashed-token").
					Return(domain.RefreshToken{UserID: userID}, nil).
					Once()

				repo.EXPECT().
					DeleteRefreshToken(mock.Anything, "hashed-token").
					Return(nil).
					Once()
			},
		},
		{
			name:     "unknown token is idempotent success",
			rawToken: "raw-token",
			setupMocks: func(
				repo *MockAuthRepository,
				refreshGen *MockRefreshTokenManager,
			) {
				refreshGen.EXPECT().
					Hash("raw-token").
					Return("hashed-token").
					Once()

				repo.EXPECT().
					GetRefreshTokenByHash(mock.Anything, "hashed-token").
					Return(domain.RefreshToken{}, errs.ErrRefreshTokenNotFound).
					Once()
			},
		},
		{
			name:     "delete error is wrapped and unwrappable via errors.Is",
			rawToken: "raw-token",
			setupMocks: func(
				repo *MockAuthRepository,
				refreshGen *MockRefreshTokenManager,
			) {
				refreshGen.EXPECT().
					Hash("raw-token").
					Return("hashed-token").
					Once()

				repo.EXPECT().
					GetRefreshTokenByHash(mock.Anything, "hashed-token").
					Return(domain.RefreshToken{UserID: userID}, nil).
					Once()

				repo.EXPECT().
					DeleteRefreshToken(mock.Anything, "hashed-token").
					Return(errs.ErrInternal).
					Once()
			},
			wantErrIs: errs.ErrInternal,
		},
		{
			name:     "repository error is wrapped and unwrappable via errors.Is",
			rawToken: "raw-token",
			setupMocks: func(
				repo *MockAuthRepository,
				refreshGen *MockRefreshTokenManager,
			) {
				refreshGen.EXPECT().
					Hash("raw-token").
					Return("hashed-token").
					Once()

				repo.EXPECT().
					GetRefreshTokenByHash(mock.Anything, "hashed-token").
					Return(domain.RefreshToken{}, errs.ErrInternal).
					Once()
			},
			wantErrIs: errs.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)

			repo := NewMockAuthRepository(t)
			refreshGen := NewMockRefreshTokenManager(t)
			tt.setupMocks(repo, refreshGen)

			svc := newTestService(repo, nil, nil, refreshGen)

			err := svc.Logout(testCtx(), tt.rawToken)

			if tt.wantErrIs != nil {
				must.Error(err)
				must.ErrorIs(err, tt.wantErrIs)

				return
			}

			must.NoError(err)
		})
	}
}
