package service

import (
	"context"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func (s *AuditService) GetLogs(
	ctx context.Context,
	filter *domain.AuditLogFilter,
) ([]domain.AuditEvent, bool, error) {
	const op = "audit.service.GetLogs"

	events, err := s.auditRepository.GetLogs(ctx, filter)
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", op, err)
	}

	hasMore := len(events) > filter.Limit
	if hasMore {
		trimmed := make([]domain.AuditEvent, 0, filter.Limit)
		trimmed = append(trimmed, events[:filter.Limit]...)

		events = trimmed
	}

	if err := resolveActorLogins(ctx, s.users, events); err != nil {
		return nil, false, fmt.Errorf("%s: %w", op, err)
	}

	return events, hasMore, nil
}
