package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/repository/postgres"
)

func (r *UsersRepository) GetUserByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.User, error) {
	const op = "users.repository.GetUserByID"

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, username, password_hash, role, created_at, updated_at
	FROM users
	WHERE id=$1`

	row := r.pool.QueryRow(ctx, query, id)

	var dbUser userModel

	if err := dbUser.Scan(row); err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return domain.User{}, fmt.Errorf(
				"%s: scan user (id=%s): %w: %v",
				op,
				id,
				errs.ErrUserNotFound,
				err,
			)
		}

		return domain.User{}, fmt.Errorf(
			"%s: scan user (id=%s): %w",
			op,
			id,
			err,
		)
	}

	user, err := dbUser.toDomain()
	if err != nil {
		return domain.User{}, fmt.Errorf(
			"%s: map user: %w",
			op,
			err,
		)
	}

	return user, nil
}
