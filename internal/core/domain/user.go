package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

func NewUser(
	id uuid.UUID,
	username string,
	passwordHash string,
	role string,
	createdAt time.Time,
	updatedAt *time.Time,
) User {
	return User{
		ID:           id,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}
