package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *AuditService) DeleteLogByID(
	ctx context.Context,
	id uuid.UUID,
) error {
	const op = "audit.service.DeleteLogByID"

	if err := s.auditRepository.DeleteLogByID(ctx, id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
