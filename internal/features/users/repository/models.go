package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
)

type userModel struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

func (u *userModel) Scan(row core_postgres_pool.Row) error {
	return row.Scan(
		&u.ID,
		&u.Username,
		&u.PasswordHash,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
}

func (u *userModel) toDomain() (domain.User, error) {
	const op = "users.repository.toDomain"

	user, err := domain.NewUser(
		u.ID,
		u.Username,
		u.PasswordHash,
		u.Role,
		u.CreatedAt,
		u.UpdatedAt,
	)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
