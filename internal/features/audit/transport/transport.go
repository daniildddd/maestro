package transport

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	core_http_server "github.com/daniildddd/maestro/internal/core/transport/server"
)

type AuditHTTPHandler struct {
	auditService AuditService
}

type AuditService interface {
	GetLogs(
		ctx context.Context,
		filter *domain.AuditLogFilter,
	) ([]domain.AuditEvent, bool, error)

	GetLogByID(
		ctx context.Context,
		id uuid.UUID,
	) (domain.AuditEvent, error)

	DeleteLogByID(
		ctx context.Context,
		id uuid.UUID,
	) error
}

func NewAuditHTTPHandler(
	auditService AuditService,
) *AuditHTTPHandler {
	return &AuditHTTPHandler{
		auditService: auditService,
	}
}

func (h *AuditHTTPHandler) PrivateRoutes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodGet,
			Path:    "/audit-logs",
			Handler: h.GetLogs,
			Roles:   []string{domain.RoleAdmin},
		},
		{
			Method:  http.MethodGet,
			Path:    "/audit-logs/{id}",
			Handler: h.GetLogByID,
			Roles:   []string{domain.RoleAdmin},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/audit-logs/{id}",
			Handler: h.DeleteLogByID,
			Roles:   []string{domain.RoleAdmin},
		},
	}
}
