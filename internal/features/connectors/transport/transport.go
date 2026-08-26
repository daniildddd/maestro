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
	GetConnectors(
		ctx context.Context,
		filter *domain.ConnectorFilter,
	) ([]domain.Connector, error)

	GetConnectorByID(
		ctx context.Context,
		name string,
	) (domain.Connector, error)

	CreateConnector(
		ctx context.Context,
		name string,
		config map[string]string,
	) (domain.Connector, error)

	PauseConnector(
		ctx context.Context,
		name string,
	) (domain.Connector, error)

	ResumeConnector(
		ctx context.Context,
		name string,
	) (domain.Connector, error)

	Delete(
		ctx context.Context,
		name string,
	) error
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
		{
			Method:  http.MethodPost,
			Path:    "/connectors",
			Handler: h.CreateConnector,
			Roles:   []string{domain.RoleUser},
		},
		{
			Method:  http.MethodGet,
			Path:    "/connectors/{id}",
			Handler: h.GetConnector,
			Roles:   []string{domain.RoleUser},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/connectors/{id}",
			Handler: h.DeleteConnector,
			Roles:   []string{domain.RoleUser},
		},
		{
			Method:  http.MethodPost,
			Path:    "/connectors/{id}/pause",
			Handler: h.PauseConnector,
			Roles:   []string{domain.RoleUser},
		},
		{
			Method:  http.MethodPost,
			Path:    "/connectors/{id}/resume",
			Handler: h.ResumeConnector,
			Roles:   []string{domain.RoleUser},
		},
	}
}
