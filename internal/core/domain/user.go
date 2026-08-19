package domain

import (
	"fmt"
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
) (User, error) {
	const op = "core.domain.NewUser"

	u := User{
		ID:           id,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}

	if err := u.Validate(); err != nil {
		return User{}, fmt.Errorf("%s: %w", op, err)
	}

	return u, nil
}

func (u User) Validate() error {
	const op = "domain.User.Validate"

	if err := ValidateUsername(u.Username); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if u.Role != RoleUser && u.Role != RoleAdmin {
		return fmt.Errorf("%s: %w", op, ErrInvalidRole)
	}

	return nil
}
