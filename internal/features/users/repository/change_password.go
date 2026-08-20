package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/errs"
)

func (r *UsersRepository) ChangePassword(
	ctx context.Context,
	id uuid.UUID,
	passwordHash string,
) error {
	const op = "users.repository.ChangePassword"

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE users
	SET password_hash=$1
	WHERE id=$2`

	tag, err := r.pool.Exec(ctx, query, passwordHash, id)
	if err != nil {
		return fmt.Errorf(
			"%s: exec query (user_id=%s): %w",
			op,
			id,
			err,
		)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf(
			"%s: user (user_id=%s): %w",
			op,
			id,
			errs.ErrUserNotFound,
		)
	}

	return nil
}
