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

type refreshTokenModel struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
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
	const op = "auth.repository.toDomain"

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

func (r *refreshTokenModel) Scan(row core_postgres_pool.Row) error {
	return row.Scan(
		&r.ID,
		&r.UserID,
		&r.TokenHash,
		&r.CreatedAt,
		&r.ExpiresAt,
	)
}

func (r *refreshTokenModel) toDomain() domain.RefreshToken {
	return domain.NewRefreshToken(
		r.ID,
		r.UserID,
		r.TokenHash,
		r.CreatedAt,
		r.ExpiresAt,
	)
}
