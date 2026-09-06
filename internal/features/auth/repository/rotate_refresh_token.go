package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
)

const maxDeadlockRetries = 3

func (r *AuthRepository) RotateRefreshToken(
	ctx context.Context,
	oldHash string,
	newToken domain.RefreshToken,
) error {
	const op = "auth.repository.RotateRefreshToken"

	for attempt := 1; ; attempt++ {
		err := r.rotateOnce(ctx, oldHash, newToken)
		if err == nil {
			return nil
		}

		if !errors.Is(err, core_postgres_pool.ErrDeadlock) {
			return fmt.Errorf("%s: %w", op, err)
		}

		if attempt == maxDeadlockRetries {
			break
		}
	}

	return fmt.Errorf("%s: retries exhausted: %w", op, core_postgres_pool.ErrDeadlock)
}

func (r *AuthRepository) rotateOnce(
	ctx context.Context,
	oldHash string,
	newToken domain.RefreshToken,
) error {
	const op = "auth.repository.rotateOnce"

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: begin tx: %w", op, err)
	}

	defer func() { _ = tx.Rollback(ctx) }() //nolint:errcheck // no-op after commit

	tag, err := tx.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE token_hash=$1`, oldHash)
	if err != nil {
		return fmt.Errorf("%s: delete old token: %w", op, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, errs.ErrRefreshTokenNotFound)
	}

	if _, err = tx.Exec(ctx,
		`INSERT INTO refresh_tokens (id, user_id, token_hash, created_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		newToken.ID, newToken.UserID, newToken.TokenHash, newToken.CreatedAt, newToken.ExpiresAt,
	); err != nil {
		return fmt.Errorf("%s: insert new token: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: commit tx: %w", op, err)
	}

	return nil
}
