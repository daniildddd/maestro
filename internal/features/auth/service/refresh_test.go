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

//nolint:maintidx // table-driven test: complexity comes from per-case mock setups
func TestRefresh(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	expiresAt := time.Now().Add(time.Hour)

	tests := []struct {
		name        string
		rawOldToken string
		setupMocks  func(
			repo *MockAuthRepository,
			accessGen *MockAccessTokenGenerator,
			refreshGen *MockRefreshTokenManager,
		)
		wantErr      bool
		wantErrIs    error
		wantNotErrIs error
		wantPair     domain.TokenPair
	}{
		{
			name:        "success rotates tokens and saves new hashed refresh token",
			rawOldToken: "old-raw",
			setupMocks: func(
				repo *MockAuthRepository,
				accessGen *MockAccessTokenGenerator,
				refreshGen *MockRefreshTokenManager,
			) {
				refreshGen.EXPECT().
					Hash("old-raw").
					Return("old-hash").
					Once()

				repo.EXPECT().
					GetRefreshTokenByHash(mock.Anything, "old-hash").
					Return(domain.NewRefreshToken(
						uuid.New(),
						userID,
						"old-hash",
						time.Now().Add(-time.Hour),
						expiresAt,
					), nil).
					Once()

				repo.EXPECT().
					GetUserByID(mock.Anything, userID).
					Return(domain.User{ID: userID, Role: "admin"}, nil).
					Once()

				repo.EXPECT().
					DeleteRefreshToken(mock.Anything, "old-hash").
					Return(nil).
					Once()

				refreshGen.EXPECT().
					Generate().
					Return("new-raw", expiresAt, nil).
					Once()

				refreshGen.EXPECT().
					Hash("new-raw").
					Return("new-hash").
					Once()

				repo.EXPECT().
					SaveRefreshToken(
						mock.Anything,
						mock.MatchedBy(func(token domain.RefreshToken) bool {
							return token.UserID == userID && token.TokenHash == "new-hash"
						}),
					).
					Return(nil).
					Once()

				accessGen.EXPECT().
					Generate(userID, "admin").
					Return("new-access", nil).
					Once()
			},
			wantPair: domain.TokenPair{
				AccessToken:  "new-access",
				RefreshToken: "new-raw",
				ExpiresAt:    expiresAt,
			},
		},
		{
			name:        "get refresh token error is wrapped",
			rawOldToken: "old-raw",
			setupMocks: func(
				repo *MockAuthRepository,
				_ *MockAccessTokenGenerator,
				refreshGen *MockRefreshTokenManager,
			) {
				refreshGen.EXPECT().
					Hash("old-raw").
					Return("old-hash").
					Once()

				repo.EXPECT().
					GetRefreshTokenByHash(mock.Anything, "old-hash").
					Return(domain.RefreshToken{}, errs.ErrInternal).
					Once()
			},
			wantErr:   true,
			wantErrIs: errs.ErrInternal,
		},
		{
			name:        "expired token returns ErrExpiredRefreshToken",
			rawOldToken: "old-raw",
			setupMocks: func(
				repo *MockAuthRepository,
				_ *MockAccessTokenGenerator,
				refreshGen *MockRefreshTokenManager,
			) {
				refreshGen.EXPECT().
					Hash("old-raw").
					Return("old-hash").
					Once()

				repo.EXPECT().
					GetRefreshTokenByHash(mock.Anything, "old-hash").
					Return(domain.NewRefreshToken(
						uuid.New(),
						userID,
						"old-hash",
						time.Now().Add(-time.Hour),
						time.Now().Add(-time.Minute),
					), nil).
					Once()
			},
			wantErr:   true,
			wantErrIs: errs.ErrExpiredRefreshToken,
		},
		{
			name:        "user not found maps to ErrInvalidRefreshToken",
			rawOldToken: "old-raw",
			setupMocks: func(
				repo *MockAuthRepository,
				_ *MockAccessTokenGenerator,
				refreshGen *MockRefreshTokenManager,
			) {
				refreshGen.EXPECT().
					Hash("old-raw").
					Return("old-hash").
					Once()

				repo.EXPECT().
					GetRefreshTokenByHash(mock.Anything, "old-hash").
					Return(domain.NewRefreshToken(
						uuid.New(),
						userID,
						"old-hash",
						time.Now().Add(-time.Hour),
						expiresAt,
					), nil).
					Once()

				repo.EXPECT().
					GetUserByID(mock.Anything, userID).
					Return(domain.User{}, errs.ErrUserNotFound).
					Once()
			},
			wantErr:      true,
			wantErrIs:    errs.ErrInvalidRefreshToken,
			wantNotErrIs: errs.ErrUserNotFound,
		},
		{
			name:        "get user by id error is wrapped",
			rawOldToken: "old-raw",
			setupMocks: func(
				repo *MockAuthRepository,
				_ *MockAccessTokenGenerator,
				refreshGen *MockRefreshTokenManager,
			) {
				refreshGen.EXPECT().
					Hash("old-raw").
					Return("old-hash").
					Once()

				repo.EXPECT().
					GetRefreshTokenByHash(mock.Anything, "old-hash").
					Return(domain.NewRefreshToken(
						uuid.New(),
						userID,
						"old-hash",
						time.Now().Add(-time.Hour),
						expiresAt,
					), nil).
					Once()

				repo.EXPECT().
					GetUserByID(mock.Anything, userID).
					Return(domain.User{}, errs.ErrInternal).
					Once()
			},
			wantErr:   true,
			wantErrIs: errs.ErrInternal,
		},
		{
			name:        "delete old refresh token error is wrapped",
			rawOldToken: "old-raw",
			setupMocks: func(
				repo *MockAuthRepository,
				_ *MockAccessTokenGenerator,
				refreshGen *MockRefreshTokenManager,
			) {
				refreshGen.EXPECT().
					Hash("old-raw").
					Return("old-hash").
					Once()

				repo.EXPECT().
					GetRefreshTokenByHash(mock.Anything, "old-hash").
					Return(domain.NewRefreshToken(
						uuid.New(),
						userID,
						"old-hash",
						time.Now().Add(-time.Hour),
						expiresAt,
					), nil).
					Once()

				repo.EXPECT().
					GetUserByID(mock.Anything, userID).
					Return(domain.User{ID: userID}, nil).
					Once()

				repo.EXPECT().
					DeleteRefreshToken(mock.Anything, "old-hash").
					Return(errs.ErrInternal).
					Once()
			},
			wantErr:   true,
			wantErrIs: errs.ErrInternal,
		},
		{
			name:        "generate new refresh token error is wrapped",
			rawOldToken: "old-raw",
			setupMocks: func(
				repo *MockAuthRepository,
				_ *MockAccessTokenGenerator,
				refreshGen *MockRefreshTokenManager,
			) {
				refreshGen.EXPECT().
					Hash("old-raw").
					Return("old-hash").
					Once()

				repo.EXPECT().
					GetRefreshTokenByHash(mock.Anything, "old-hash").
					Return(domain.NewRefreshToken(
						uuid.New(),
						userID,
						"old-hash",
						time.Now().Add(-time.Hour),
						expiresAt,
					), nil).
					Once()

				repo.EXPECT().
					GetUserByID(mock.Anything, userID).
					Return(domain.User{ID: userID}, nil).
					Once()

				repo.EXPECT().
					DeleteRefreshToken(mock.Anything, "old-hash").
					Return(nil).
					Once()

				refreshGen.EXPECT().
					Generate().
					Return("", time.Time{}, errs.ErrInternal).
					Once()
			},
			wantErr:   true,
			wantErrIs: errs.ErrInternal,
		},
		{
			name:        "create new refresh token error is wrapped",
			rawOldToken: "old-raw",
			setupMocks: func(
				repo *MockAuthRepository,
				_ *MockAccessTokenGenerator,
				refreshGen *MockRefreshTokenManager,
			) {
				refreshGen.EXPECT().
					Hash("old-raw").
					Return("old-hash").
					Once()

				repo.EXPECT().
					GetRefreshTokenByHash(mock.Anything, "old-hash").
					Return(domain.NewRefreshToken(
						uuid.New(),
						userID,
						"old-hash",
						time.Now().Add(-time.Hour),
						expiresAt,
					), nil).
					Once()

				repo.EXPECT().
					GetUserByID(mock.Anything, userID).
					Return(domain.User{ID: userID}, nil).
					Once()

				repo.EXPECT().
					DeleteRefreshToken(mock.Anything, "old-hash").
					Return(nil).
					Once()

				refreshGen.EXPECT().
					Generate().
					Return("new-raw", time.Now().Add(-time.Hour), nil).
					Once()

				refreshGen.EXPECT().
					Hash("new-raw").
					Return("new-hash").
					Once()
			},
			wantErr: true,
		},
		{
			name:        "save new refresh token error is wrapped",
			rawOldToken: "old-raw",
			setupMocks: func(
				repo *MockAuthRepository,
				_ *MockAccessTokenGenerator,
				refreshGen *MockRefreshTokenManager,
			) {
				refreshGen.EXPECT().
					Hash("old-raw").
					Return("old-hash").
					Once()

				repo.EXPECT().
					GetRefreshTokenByHash(mock.Anything, "old-hash").
					Return(domain.NewRefreshToken(
						uuid.New(),
						userID,
						"old-hash",
						time.Now().Add(-time.Hour),
						expiresAt,
					), nil).
					Once()

				repo.EXPECT().
					GetUserByID(mock.Anything, userID).
					Return(domain.User{ID: userID}, nil).
					Once()

				repo.EXPECT().
					DeleteRefreshToken(mock.Anything, "old-hash").
					Return(nil).
					Once()

				refreshGen.EXPECT().
					Generate().
					Return("new-raw", expiresAt, nil).
					Once()

				refreshGen.EXPECT().
					Hash("new-raw").
					Return("new-hash").
					Once()

				repo.EXPECT().
					SaveRefreshToken(mock.Anything, mock.Anything).
					Return(errs.ErrInternal).
					Once()
			},
			wantErr:   true,
			wantErrIs: errs.ErrInternal,
		},
		{
			name:        "generate access token error is wrapped",
			rawOldToken: "old-raw",
			setupMocks: func(
				repo *MockAuthRepository,
				accessGen *MockAccessTokenGenerator,
				refreshGen *MockRefreshTokenManager,
			) {
				refreshGen.EXPECT().
					Hash("old-raw").
					Return("old-hash").
					Once()

				repo.EXPECT().
					GetRefreshTokenByHash(mock.Anything, "old-hash").
					Return(domain.NewRefreshToken(
						uuid.New(),
						userID,
						"old-hash",
						time.Now().Add(-time.Hour),
						expiresAt,
					), nil).
					Once()

				repo.EXPECT().
					GetUserByID(mock.Anything, userID).
					Return(domain.User{ID: userID, Role: "admin"}, nil).
					Once()

				repo.EXPECT().
					DeleteRefreshToken(mock.Anything, "old-hash").
					Return(nil).
					Once()

				refreshGen.EXPECT().
					Generate().
					Return("new-raw", expiresAt, nil).
					Once()

				refreshGen.EXPECT().
					Hash("new-raw").
					Return("new-hash").
					Once()

				repo.EXPECT().
					SaveRefreshToken(mock.Anything, mock.Anything).
					Return(nil).
					Once()

				accessGen.EXPECT().
					Generate(userID, "admin").
					Return("", errs.ErrInternal).
					Once()
			},
			wantErr:   true,
			wantErrIs: errs.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			repo := NewMockAuthRepository(t)
			accessGen := NewMockAccessTokenGenerator(t)
			refreshGen := NewMockRefreshTokenManager(t)
			tt.setupMocks(repo, accessGen, refreshGen)

			svc := newTestService(repo, nil, accessGen, refreshGen)

			pair, err := svc.Refresh(context.Background(), tt.rawOldToken)

			if tt.wantErr {
				must.Error(err)

				if tt.wantErrIs != nil {
					must.ErrorIs(err, tt.wantErrIs)
				}

				if tt.wantNotErrIs != nil {
					must.NotErrorIs(err, tt.wantNotErrIs)
				}

				return
			}

			must.NoError(err)
			is.Equal(tt.wantPair, pair)
		})
	}
}
