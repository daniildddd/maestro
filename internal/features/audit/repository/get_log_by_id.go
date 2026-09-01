package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/repository/postgres"
)

func (r *AuditRepository) GetLogByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.AuditEvent, error) {
	const op = "audit.repository.GetLogByID"

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, action, outcome, failure_reason, actor_id, actor_login,
	       subject_type, subject_id, subject_name, state_before, state_after,
	       request_id, ip, user_agent, created_at
	FROM audit_logs
	WHERE id=$1`

	row := r.pool.QueryRow(ctx, query, id)

	var dbEvent auditEventModel

	if err := dbEvent.Scan(row); err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return domain.AuditEvent{}, fmt.Errorf(
				"%s: scan event (id=%s): %w: %v",
				op,
				id,
				errs.ErrAuditLogNotFound,
				err,
			)
		}

		return domain.AuditEvent{}, fmt.Errorf(
			"%s: scan event (id=%s): %w",
			op,
			id,
			err,
		)
	}

	event, err := dbEvent.toDomain()
	if err != nil {
		return domain.AuditEvent{}, fmt.Errorf(
			"%s: map event: %w",
			op,
			err,
		)
	}

	return event, nil
}
