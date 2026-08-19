package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
)

func (r *UsersRepository) CreateUser(
	ctx context.Context,
	user domain.User,
) (domain.User, error) {
	const op = "users.repository.CreateUser"

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	INSERT INTO users (id, username, password_hash, role, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, username, password_hash, role, created_at, updated_at`

	row := r.pool.QueryRow(
		ctx,
		query,
		user.ID,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)

	var dbUser userModel

	if err := dbUser.Scan(row); err != nil {
		if errors.Is(err, core_postgres_pool.ErrDuplicate) {
			return domain.User{}, fmt.Errorf(
				"%s: insert user (username=%s): %w: %v",
				op,
				user.Username,
				errs.ErrUsernameConflict,
				err,
			)
		}

		return domain.User{}, fmt.Errorf(
			"%s: scan row (username=%s): %w",
			op,
			user.Username,
			err,
		)
	}

	created, err := dbUser.toDomain()
	if err != nil {
		return domain.User{}, fmt.Errorf(
			"%s: map user: %w",
			op,
			err,
		)
	}

	return created, nil
}
