package domain

import (
	"errors"
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
	if userId == uuid.Nil {
		return RefreshToken{}, errors.New("userId must not be nil")
	}

	if token == "" {
		return RefreshToken{}, errors.New("rawToken must not be empty")
	}

	if !expiresAt.After(time.Now()) {
		return RefreshToken{}, errors.New("expiresAt must be after now")
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
