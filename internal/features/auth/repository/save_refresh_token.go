package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/repository/postgres"
)

func (r *AuthRepository) SaveRefreshToken(
	ctx context.Context,
	token domain.RefreshToken,
) error {
	const op = "auth.repository.SaveRefreshToken"

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	INSERT INTO refresh_tokens (id, user_id, token_hash, created_at, expires_at)
	VALUES ($1, $2, $3, $4, $5)`

	_, err := r.pool.Exec(
		ctx,
		query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.CreatedAt,
		token.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, postgres.ErrViolatesForeignKey) {
			return fmt.Errorf("%s: save refresh token for user_id=%s: %w: %v",
				op,
				token.UserID,
				postgres.ErrViolatesForeignKey,
				err,
			)
		}

		return fmt.Errorf("%s: exec query: %w",
			op,
			err,
		)
	}

	return nil
}
