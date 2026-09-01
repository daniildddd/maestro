package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
)

type AuditService struct {
	auditRepository AuditRepository
}

func NewAuditService(
	auditRepository AuditRepository,
) *AuditService {
	return &AuditService{
		auditRepository: auditRepository,
	}
}

type AuditRepository interface {
	GetLogs(
		ctx context.Context,
		filter *domain.AuditLogFilter,
	) ([]domain.AuditEvent, error)

	GetLogByID(
		ctx context.Context,
		id uuid.UUID,
	) (domain.AuditEvent, error)
}
