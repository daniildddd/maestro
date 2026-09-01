package transport

import (
	"context"
	"net/http"

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
	}
}
