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

func (r *UsersRepository) UpdateUser(
	ctx context.Context,
	id uuid.UUID,
	username string,
) (domain.User, error) {
	const op = "users.repository.UpdateUser"

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE users
	SET username=$1, updated_at=NOW()
	WHERE id=$2
	RETURNING id, username, password_hash, role, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query, username, id)

	var dbUser userModel

	if err := dbUser.Scan(row); err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return domain.User{}, fmt.Errorf(
				"%s: user not found (user_id=%s): %w: %v",
				op,
				id,
				errs.ErrUserNotFound,
				err,
			)
		}

		return domain.User{}, fmt.Errorf(
			"%s: scan row (user_id=%s): %w",
			op,
			id,
			err,
		)
	}

	user, err := dbUser.toDomain()
	if err != nil {
		return domain.User{}, fmt.Errorf(
			"%s: map user (user_id=%s): %w",
			id,
			op,
			err,
		)
	}

	return user, nil
}
