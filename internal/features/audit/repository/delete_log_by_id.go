package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/errs"
)

func (r *AuditRepository) DeleteLogByID(
	ctx context.Context,
	id uuid.UUID,
) error {
	const op = "audit.repository.DeleteLogByID"

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	DELETE FROM audit_logs
	WHERE id=$1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf(
			"%s: exec query (id=%s): %w",
			op,
			id,
			err,
		)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf(
			"%s: log not found (id=%s): %w",
			op,
			id,
			errs.ErrAuditLogNotFound,
		)
	}

	return nil
}
