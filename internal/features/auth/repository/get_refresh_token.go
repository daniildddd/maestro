package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/repository/postgres"
)

func (r *AuthRepository) GetRefreshTokenByHash(
	ctx context.Context,
	tokenHash string,
) (domain.RefreshToken, error) {
	const op = "auth.repository.GetRefreshTokenByHash"

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, user_id, token_hash, created_at, expires_at
	FROM refresh_tokens
	WHERE token_hash=$1`

	row := r.pool.QueryRow(ctx, query, tokenHash)

	var dbToken refreshTokenModel

	if err := dbToken.Scan(row); err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return domain.RefreshToken{}, fmt.Errorf(
				"%s: scan row: %w: %v",
				op,
				errs.ErrRefreshTokenNotFound,
				err,
			)
		}

		return domain.RefreshToken{}, fmt.Errorf(
			"%s: scan row: %w",
			op,
			err,
		)
	}

	return dbToken.toDomain(), nil
}
