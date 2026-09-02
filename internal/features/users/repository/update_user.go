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

func (r *UsersRepository) UpdateUser(
	ctx context.Context,
	id uuid.UUID,
	username string,
) (before, after domain.User, err error) {
	const op = "users.repository.UpdateUser"

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	WITH old AS (
		SELECT id, username, password_hash, role, created_at, updated_at
		FROM users
		WHERE id=$2
	), upd AS (
		UPDATE users
		SET username=$1, updated_at=NOW()
		WHERE id=$2
		RETURNING id, username, password_hash, role, created_at, updated_at
	)
	SELECT
		old.id, old.username, old.password_hash, old.role, old.created_at, old.updated_at,
		upd.id, upd.username, upd.password_hash, upd.role, upd.created_at, upd.updated_at
	FROM old, upd`

	row := r.pool.QueryRow(ctx, query, username, id)

	var oldUser userModel

	var updatedUser userModel

	if err := scanUserPair(row, &oldUser, &updatedUser); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.User{}, domain.User{}, fmt.Errorf(
				"%s: user not found (user_id=%s): %w: %v",
				op,
				id,
				errs.ErrUserNotFound,
				err,
			)
		}

		if errors.Is(err, core_postgres_pool.ErrDuplicate) {
			return domain.User{}, domain.User{}, fmt.Errorf(
				"%s: update user (user_id=%s, username=%s): %w: %v",
				op,
				id,
				username,
				errs.ErrUsernameConflict,
				err,
			)
		}

		return domain.User{}, domain.User{}, fmt.Errorf(
			"%s: scan row (user_id=%s): %w",
			op,
			id,
			err,
		)
	}

	before, err = oldUser.toDomain()
	if err != nil {
		return domain.User{}, domain.User{}, fmt.Errorf(
			"%s: map user (user_id=%s): %w",
			op,
			id,
			err,
		)
	}

	after, err = updatedUser.toDomain()
	if err != nil {
		return domain.User{}, domain.User{}, fmt.Errorf(
			"%s: map user (user_id=%s): %w",
			op,
			id,
			err,
		)
	}

	return before, after, nil
}

func scanUserPair(
	row core_postgres_pool.Row,
	old *userModel,
	updated *userModel,
) error {
	return row.Scan(
		&old.ID,
		&old.Username,
		&old.PasswordHash,
		&old.Role,
		&old.CreatedAt,
		&old.UpdatedAt,
		&updated.ID,
		&updated.Username,
		&updated.PasswordHash,
		&updated.Role,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
}
