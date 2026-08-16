package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
)

func (r *AuthRepository) GetUserByName(
	ctx context.Context,
	username string,
) (domain.User, error) {
	const op = "auth.repository.GetUserByName"

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, username, password_hash, role, created_at, updated_at
	FROM users
	WHERE username=$1;`

	row := r.pool.QueryRow(ctx, query, username)

	var dbUser userModel

	if err := dbUser.Scan(row); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.User{}, fmt.Errorf(
				"%s: scan row (username=%s): %w: %v",
				op,
				username,
				errs.ErrUserNotFound,
				err,
			)
		}

		return domain.User{}, fmt.Errorf(
			"%s: scan row (username=%s): %w",
			op,
			username,
			err,
		)
	}

	return dbUser.toDomain(), nil
}
