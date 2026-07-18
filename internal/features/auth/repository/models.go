package repository

import (
	"time"

	"github.com/daniildddd/maestro/internal/core/domain"
	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
	"github.com/google/uuid"
)

type UserModel struct {
	Id           uuid.UUID
	Username     string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

func (u *UserModel) Scan(row core_postgres_pool.Row) error {
	return row.Scan(
		&u.Id,
		&u.Username,
		&u.PasswordHash,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
}

func userModelToDomain(model UserModel) domain.User {
	return domain.NewUser(
		model.Id,
		model.Username,
		model.PasswordHash,
		model.Role,
		model.CreatedAt,
		model.UpdatedAt,
	)
}
