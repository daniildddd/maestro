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
	RevokedAt *time.Time
}

func CreateRefreshToken(
	userId uuid.UUID,
	token string,
	expiresAt time.Time,
) (RefreshToken, error) {
	if userId == uuid.Nil {
		return RefreshToken{}, fmt.Errorf("userId must not be nil")
	}

	if token == "" {
		return RefreshToken{}, fmt.Errorf("rawToken must not be empty")
	}

	if !expiresAt.After(time.Now()) {
		return RefreshToken{}, fmt.Errorf("expiresAt must be after now")
	}

	return RefreshToken{
		Id:        uuid.New(),
		UserId:    userId,
		TokenHash: token,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		RevokedAt: nil,
	}, nil
}
