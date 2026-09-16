package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
)

type AuditService struct {
	auditRepository AuditRepository
	users           UserDirectory
}

func NewAuditService(
	auditRepository AuditRepository,
	users UserDirectory,
) *AuditService {
	return &AuditService{
		auditRepository: auditRepository,
		users:           users,
	}
}

type UserDirectory interface {
	GetUsersByIDs(
		ctx context.Context,
		ids []uuid.UUID,
	) ([]domain.User, error)
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

	DeleteLogByID(
		ctx context.Context,
		id uuid.UUID,
	) error
}
