package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func (r *AuditRepository) Record(
	ctx context.Context,
	event domain.AuditEvent,
) {
	const op = "audit.repository.Record"

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	INSERT INTO audit_logs (
		id, action, outcome, failure_reason,
		actor_id, actor_login,
		subject_type, subject_id, subject_name,
		state_before, state_after,
		request_id, ip, user_agent,
		created_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	stateBefore, err := marshalState(event.StateBefore)
	if err != nil {
		r.logger.Error(
			"record audit event",
			zap.String("action", string(event.Action)),
			zap.String("outcome", string(event.Outcome)),
			zap.Error(fmt.Errorf("%s: marshal state_before: %w", op, err)),
		)

		return
	}

	stateAfter, err := marshalState(event.StateAfter)
	if err != nil {
		r.logger.Error(
			"record audit event",
			zap.String("action", string(event.Action)),
			zap.String("outcome", string(event.Outcome)),
			zap.Error(fmt.Errorf("%s: marshal state_after: %w", op, err)),
		)

		return
	}

	_, err = r.pool.Exec(
		ctx,
		query,
		event.ID,
		string(event.Action),
		string(event.Outcome),
		nullableString(event.FailureReason),
		nullableUUID(event.ActorID),
		nullableString(event.ActorLogin),
		event.SubjectType,
		nullableString(event.SubjectID),
		nullableString(event.SubjectName),
		stateBefore,
		stateAfter,
		nullableString(event.RequestID),
		nullableString(event.IP),
		nullableString(event.UserAgent),
		event.CreatedAt,
	)
	if err != nil {
		r.logger.Error(
			"record audit event",
			zap.String("action", string(event.Action)),
			zap.String("outcome", string(event.Outcome)),
			zap.Error(fmt.Errorf("%s: exec query: %w", op, err)),
		)

		return
	}
}

func marshalState(state map[string]any) ([]byte, error) {
	if state == nil {
		return nil, nil
	}

	return json.Marshal(state)
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

func nullableUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}

	return &id
}
