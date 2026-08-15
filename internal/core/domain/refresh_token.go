package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func CreateRefreshToken(
	userId uuid.UUID,
	token string,
	expiresAt time.Time,
) (RefreshToken, error) {
	const op = "core.domain.CreateRefreshToken"

	if userId == uuid.Nil {
		return RefreshToken{}, fmt.Errorf("%s: userId must not be nil", op)
	}

	if token == "" {
		return RefreshToken{}, fmt.Errorf("%s: rawToken must not be empty", op)
	}

	if !expiresAt.After(time.Now()) {
		return RefreshToken{}, fmt.Errorf("%s: expiresAt must be after now", op)
	}

	return RefreshToken{
		Id:        uuid.New(),
		UserId:    userId,
		TokenHash: token,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}, nil
}

func NewRefreshToken(
	id uuid.UUID,
	userId uuid.UUID,
	tokenHash string,
	createdAt time.Time,
	expiresAt time.Time,
) RefreshToken {
	return RefreshToken{
		Id:        id,
		UserId:    userId,
		TokenHash: tokenHash,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
	}
}

func (t *RefreshToken) IsExpired(now time.Time) bool {
	return t.ExpiresAt.Before(now)
}
