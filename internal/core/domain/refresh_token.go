package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func CreateRefreshToken(
	userID uuid.UUID,
	token string,
	expiresAt time.Time,
) (RefreshToken, error) {
	const op = "core.domain.CreateRefreshToken"

	if userID == uuid.Nil {
		return RefreshToken{}, fmt.Errorf("%s: userID must not be nil", op)
	}

	if token == "" {
		return RefreshToken{}, fmt.Errorf("%s: rawToken must not be empty", op)
	}

	if !expiresAt.After(time.Now()) {
		return RefreshToken{}, fmt.Errorf("%s: expiresAt must be after now", op)
	}

	return RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: token,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}, nil
}

func NewRefreshToken(
	id uuid.UUID,
	userID uuid.UUID,
	tokenHash string,
	createdAt time.Time,
	expiresAt time.Time,
) RefreshToken {
	return RefreshToken{
		ID:        id,
		UserID:    userID,
		TokenHash: tokenHash,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
	}
}

func (t *RefreshToken) IsExpired(now time.Time) bool {
	return t.ExpiresAt.Before(now)
}
