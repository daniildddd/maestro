package repository

import (
	"context"
	"fmt"
)

func (r *AuthRepository) DeleteRefreshToken(
	ctx context.Context,
	tokenHash string,
) error {
	const op = "auth.repository.DeleteRefreshToken"

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	DELETE FROM refresh_tokens
	WHERE token_hash=$1`

	_, err := r.pool.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf(
			"%s: exec query: %w",
			op,
			err,
		)
	}

	return nil
}
