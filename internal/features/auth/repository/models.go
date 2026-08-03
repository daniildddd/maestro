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

type refreshTokenModel struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
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

func (r *refreshTokenModel) Scan(row core_postgres_pool.Row) error {
	return row.Scan(
		&r.Id,
		&r.UserId,
		&r.TokenHash,
		&r.CreatedAt,
		&r.ExpiresAt,
	)
}

func (r *refreshTokenModel) toDomain() domain.RefreshToken {
	return domain.NewRefreshToken(
		r.Id,
		r.UserId,
		r.TokenHash,
		r.CreatedAt,
		r.ExpiresAt,
	)
}