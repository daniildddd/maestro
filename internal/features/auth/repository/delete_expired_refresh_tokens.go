package repository

import (
	"context"
	"fmt"
)

func (r *AuthRepository) DeleteExpiredRefreshTokens(
	ctx context.Context,
) (int64, error) {
	const op = "auth.repository.DeleteExpiredRefreshTokens"

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tag, err := r.pool.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE expires_at < now()`)
	if err != nil {
		return 0, fmt.Errorf("%s: exec query: %w", op, err)
	}

	return tag.RowsAffected(), nil
}
