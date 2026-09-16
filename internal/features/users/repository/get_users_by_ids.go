package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func (r *UsersRepository) GetUsersByIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]domain.User, error) {
	const op = "users.repository.GetUsersByIDs"

	if len(ids) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	rows, err := r.pool.Query(ctx, `
	SELECT id, username, password_hash, role, created_at, updated_at
	FROM users WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, fmt.Errorf(
			"%s: query: %w",
			op,
			err,
		)
	}
	defer rows.Close()

	var users []domain.User

	for rows.Next() {
		var dbUser userModel

		if err := dbUser.Scan(rows); err != nil {
			return nil, fmt.Errorf(
				"%s: scan row: %w",
				op,
				err,
			)
		}

		user, err := dbUser.toDomain()
		if err != nil {
			return nil, fmt.Errorf(
				"%s: map user: %w",
				op,
				err,
			)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"%s: iterate rows: %w",
			op,
			err,
		)
	}

	return users, nil
}
