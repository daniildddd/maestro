package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func (s *AuditService) GetLogByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.AuditEvent, error) {
	const op = "audit.service.GetLogByID"

	event, err := s.auditRepository.GetLogByID(ctx, id)
	if err != nil {
		return domain.AuditEvent{}, fmt.Errorf("%s: %w", op, err)
	}

	events := []domain.AuditEvent{event}

	if err := resolveActorLogins(ctx, s.users, events); err != nil {
		return domain.AuditEvent{}, fmt.Errorf("%s: %w", op, err)
	}

	return events[0], nil
}
