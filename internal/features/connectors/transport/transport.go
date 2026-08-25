package transport

import (
	"context"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/domain"
	core_http_server "github.com/daniildddd/maestro/internal/core/transport/server"
)

type ConnectorsHTTPHandler struct {
	connectorsService ConnectorsService
}

type ConnectorsService interface {
	List(
		ctx context.Context,
		filter *domain.ConnectorFilter,
	) ([]domain.Connector, error)
}

func NewConnectorsHTTPHandler(connectorsService ConnectorsService) *ConnectorsHTTPHandler {
	return &ConnectorsHTTPHandler{
		connectorsService: connectorsService,
	}
}

func (h *ConnectorsHTTPHandler) PrivateRoutes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodGet,
			Path:    "/connectors",
			Handler: h.GetConnectors,
			Roles:   []string{domain.RoleUser},
		},
	}
}
