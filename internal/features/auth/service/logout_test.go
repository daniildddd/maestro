package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

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
		wantErr bool
	}{
		{
			name:     "success deletes token by hash",
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
					DeleteRefreshToken(mock.Anything, "hashed-token").
					Return(nil).
					Once()
			},
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
					DeleteRefreshToken(mock.Anything, "hashed-token").
					Return(errs.ErrInternal).
					Once()
			},
			wantErr: true,
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

			err := svc.Logout(context.Background(), tt.rawToken)

			if tt.wantErr {
				must.Error(err)
				must.ErrorIs(err, errs.ErrInternal)

				return
			}

			must.NoError(err)
		})
	}
}
