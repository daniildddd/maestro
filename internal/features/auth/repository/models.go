package repository

import (
	"time"

	"github.com/daniildddd/maestro/internal/core/domain"
	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
	"github.com/google/uuid"
)

type userModel struct {
	Id           uuid.UUID
	Username     string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

func (u *userModel) Scan(row core_postgres_pool.Row) error {
	return row.Scan(
		&u.Id,
		&u.Username,
		&u.PasswordHash,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
}

func (u *userModel) toDomain() domain.User {
	return domain.NewUser(
		u.Id,
		u.Username,
		u.PasswordHash,
		u.Role,
		u.CreatedAt,
		u.UpdatedAt,
	)
}