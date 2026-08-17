package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func TestCreateRefreshToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		userID     uuid.UUID
		token      string
		expiresAt  time.Time
		wantErrMsg string
	}{
		{
			name:       "valid inputs returns token",
			userID:     uuid.New(),
			token:      "validTokenHash",
			expiresAt:  time.Now().Add(time.Hour),
			wantErrMsg: "",
		},
		{
			name:       "nil user id rejected",
			userID:     uuid.Nil,
			token:      "validTokenHash",
			expiresAt:  time.Now().Add(time.Hour),
			wantErrMsg: "userID must not be nil",
		},
		{
			name:       "empty token rejected",
			userID:     uuid.New(),
			token:      "",
			expiresAt:  time.Now().Add(time.Hour),
			wantErrMsg: "rawToken must not be empty",
		},
		{
			name:       "past expiry rejected",
			userID:     uuid.New(),
			token:      "validTokenHash",
			expiresAt:  time.Now().Add(-time.Hour),
			wantErrMsg: "expiresAt must be after now",
		},
		{
			name:       "now expiry rejected",
			userID:     uuid.New(),
			token:      "validTokenHash",
			expiresAt:  time.Now(),
			wantErrMsg: "expiresAt must be after now",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			must := require.New(t)

			got, err := domain.CreateRefreshToken(tt.userID, tt.token, tt.expiresAt)

			if tt.wantErrMsg != "" {
				must.Error(err)
				is.Equal(domain.RefreshToken{}, got)
				is.ErrorContains(err, tt.wantErrMsg)

				return
			}

			must.NoError(err)

			is.Equal(tt.userID, got.UserID)
			is.Equal(tt.token, got.TokenHash)
			is.True(got.ExpiresAt.Equal(tt.expiresAt))
			is.NotEqual(uuid.Nil, got.ID)
			is.True(got.CreatedAt.Before(time.Now().Add(time.Second)))
		})
	}
}

func TestNewRefreshToken(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	id := uuid.New()
	userID := uuid.New()
	tokenHash := "valid-hash"
	createdAt := time.Now()
	expiresAt := time.Now().Add(time.Hour)

	got := domain.NewRefreshToken(
		id,
		userID,
		tokenHash,
		createdAt,
		expiresAt,
	)

	is.Equal(id, got.ID)
	is.Equal(userID, got.UserID)
	is.Equal(tokenHash, got.TokenHash)
	is.True(createdAt.Equal(got.CreatedAt))
	is.True(expiresAt.Equal(got.ExpiresAt))
}

func TestRefreshToken_IsExpired(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name      string
		now       time.Time
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "expired in past",
			now:       now,
			expiresAt: now.Add(-time.Hour),
			want:      true,
		},
		{
			name:      "expires in future",
			now:       now,
			expiresAt: now.Add(time.Hour),
			want:      false,
		},
		{
			name:      "exactly now not expired",
			now:       now,
			expiresAt: now,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)

			refresh := domain.RefreshToken{ExpiresAt: tt.expiresAt}

			got := refresh.IsExpired(now)

			is.Equal(tt.want, got)
		})
	}
}
