package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
)

func (r *UsersRepository) DeleteUser(
	ctx context.Context,
	id uuid.UUID,
) (deleted domain.User, err error) {
	const op = "users.repository.DeleteUser"

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	DELETE FROM users
	WHERE id=$1
	RETURNING id, username, password_hash, role, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query, id)

	var dbUser userModel

	if err := dbUser.Scan(row); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.User{}, fmt.Errorf(
				"%s: user not found (id=%s): %w: %v",
				op,
				id,
				errs.ErrUserNotFound,
				err,
			)
		}

		return domain.User{}, fmt.Errorf(
			"%s: scan row (id=%s): %w",
			op,
			id,
			err,
		)
	}

	deleted, err = dbUser.toDomain()
	if err != nil {
		return domain.User{}, fmt.Errorf(
			"%s: map user (id=%s): %w",
			op,
			id,
			err,
		)
	}

	return deleted, nil
}
