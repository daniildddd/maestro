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
func TestLogin(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	expiresAt := time.Now().Add(time.Hour)

	tests := []struct {
		name       string
		username   string
		password   string
		setupMocks func(
			repo *MockAuthRepository,
			hasher *MockPasswordHasher,
			accessGen *MockAccessTokenGenerator,
			refreshGen *MockRefreshTokenManager,
		)
		wantErr      bool
		wantErrIs    error
		wantNotErrIs error
		wantPair     domain.TokenPair
	}{
		{
			name:     "success returns token pair and saves hashed refresh token",
			username: "alice",
			password: "secret123",
			setupMocks: func(
				repo *MockAuthRepository,
				hasher *MockPasswordHasher,
				accessGen *MockAccessTokenGenerator,
				refreshGen *MockRefreshTokenManager,
			) {
				repo.EXPECT().
					GetUserByName(mock.Anything, "alice").
					Return(domain.User{ID: userID, PasswordHash: "hash", Role: "admin"}, nil).
					Once()

				hasher.EXPECT().
					Verify("hash", "secret123").
					Return(nil).
					Once()

				accessGen.EXPECT().
					Generate(userID, "admin").
					Return("access-token", nil).
					Once()

				refreshGen.EXPECT().
					Generate().
					Return("raw-refresh", expiresAt, nil).
					Once()

				refreshGen.EXPECT().
					Hash("raw-refresh").
					Return("hashed-refresh").
					Once()

				repo.EXPECT().
					SaveRefreshToken(
						mock.Anything,
						mock.MatchedBy(func(token domain.RefreshToken) bool {
							return token.UserID == userID && token.TokenHash == "hashed-refresh"
						}),
					).
					Return(nil).
					Once()
			},
			wantPair: domain.TokenPair{
				AccessToken:  "access-token",
				RefreshToken: "raw-refresh",
				ExpiresAt:    expiresAt,
			},
		},
		{
			name:     "user not found maps to ErrInvalidCredentials",
			username: "alice",
			password: "secret123",
			setupMocks: func(
				repo *MockAuthRepository,
				_ *MockPasswordHasher,
				_ *MockAccessTokenGenerator,
				_ *MockRefreshTokenManager,
			) {
				repo.EXPECT().
					GetUserByName(mock.Anything, "alice").
					Return(domain.User{}, errs.ErrUserNotFound).
					Once()
			},
			wantErr:      true,
			wantErrIs:    errs.ErrInvalidCredentials,
			wantNotErrIs: errs.ErrUserNotFound,
		},
		{
			name:     "non user-not-found error is not mapped to ErrInvalidCredentials",
			username: "alice",
			password: "secret123",
			setupMocks: func(
				repo *MockAuthRepository,
				_ *MockPasswordHasher,
				_ *MockAccessTokenGenerator,
				_ *MockRefreshTokenManager,
			) {
				repo.EXPECT().
					GetUserByName(mock.Anything, "alice").
					Return(domain.User{}, errs.ErrInternal).
					Once()
			},
			wantErr:      true,
			wantErrIs:    errs.ErrInternal,
			wantNotErrIs: errs.ErrInvalidCredentials,
		},
		{
			name:     "password verification error maps to ErrInvalidCredentials",
			username: "alice",
			password: "wrong-password",
			setupMocks: func(
				repo *MockAuthRepository,
				hasher *MockPasswordHasher,
				_ *MockAccessTokenGenerator,
				_ *MockRefreshTokenManager,
			) {
				repo.EXPECT().
					GetUserByName(mock.Anything, "alice").
					Return(domain.User{ID: userID, PasswordHash: "hash"}, nil).
					Once()

				hasher.EXPECT().
					Verify("hash", "wrong-password").
					Return(errs.ErrInternal).
					Once()
			},
			wantErr:   true,
			wantErrIs: errs.ErrInvalidCredentials,
		},
		{
			name:     "access token generation error is wrapped",
			username: "alice",
			password: "secret123",
			setupMocks: func(
				repo *MockAuthRepository,
				hasher *MockPasswordHasher,
				accessGen *MockAccessTokenGenerator,
				_ *MockRefreshTokenManager,
			) {
				repo.EXPECT().
					GetUserByName(mock.Anything, "alice").
					Return(domain.User{ID: userID, PasswordHash: "hash", Role: "admin"}, nil).
					Once()

				hasher.EXPECT().
					Verify("hash", "secret123").
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
		{
			name:     "refresh token generation error is wrapped",
			username: "alice",
			password: "secret123",
			setupMocks: func(
				repo *MockAuthRepository,
				hasher *MockPasswordHasher,
				accessGen *MockAccessTokenGenerator,
				refreshGen *MockRefreshTokenManager,
			) {
				repo.EXPECT().
					GetUserByName(mock.Anything, "alice").
					Return(domain.User{ID: userID, PasswordHash: "hash", Role: "admin"}, nil).
					Once()

				hasher.EXPECT().
					Verify("hash", "secret123").
					Return(nil).
					Once()

				accessGen.EXPECT().
					Generate(userID, "admin").
					Return("access-token", nil).
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
			name:     "create refresh token error is wrapped",
			username: "alice",
			password: "secret123",
			setupMocks: func(
				repo *MockAuthRepository,
				hasher *MockPasswordHasher,
				accessGen *MockAccessTokenGenerator,
				refreshGen *MockRefreshTokenManager,
			) {
				repo.EXPECT().
					GetUserByName(mock.Anything, "alice").
					Return(domain.User{ID: userID, PasswordHash: "hash", Role: "admin"}, nil).
					Once()

				hasher.EXPECT().
					Verify("hash", "secret123").
					Return(nil).
					Once()

				accessGen.EXPECT().
					Generate(userID, "admin").
					Return("access-token", nil).
					Once()

				refreshGen.EXPECT().
					Generate().
					Return("raw-refresh", time.Now().Add(-time.Hour), nil).
					Once()

				refreshGen.EXPECT().
					Hash("raw-refresh").
					Return("hashed-refresh").
					Once()
			},
			wantErr: true,
		},
		{
			name:     "save refresh token error is wrapped",
			username: "alice",
			password: "secret123",
			setupMocks: func(
				repo *MockAuthRepository,
				hasher *MockPasswordHasher,
				accessGen *MockAccessTokenGenerator,
				refreshGen *MockRefreshTokenManager,
			) {
				repo.EXPECT().
					GetUserByName(mock.Anything, "alice").
					Return(domain.User{ID: userID, PasswordHash: "hash", Role: "admin"}, nil).
					Once()

				hasher.EXPECT().
					Verify("hash", "secret123").
					Return(nil).
					Once()

				accessGen.EXPECT().
					Generate(userID, "admin").
					Return("access-token", nil).
					Once()

				refreshGen.EXPECT().
					Generate().
					Return("raw-refresh", expiresAt, nil).
					Once()

				refreshGen.EXPECT().
					Hash("raw-refresh").
					Return("hashed-refresh").
					Once()

				repo.EXPECT().
					SaveRefreshToken(mock.Anything, mock.Anything).
					Return(errs.ErrInternal).
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
			hasher := NewMockPasswordHasher(t)
			accessGen := NewMockAccessTokenGenerator(t)
			refreshGen := NewMockRefreshTokenManager(t)
			tt.setupMocks(repo, hasher, accessGen, refreshGen)

			svc := newTestService(repo, hasher, accessGen, refreshGen)

			pair, err := svc.Login(context.Background(), tt.username, tt.password)

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
