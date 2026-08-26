package transport

import (
	"context"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/domain"
	core_http_server "github.com/daniildddd/maestro/internal/core/transport/server"
)

const connectorsByIDPath = "/connectors/{id}"

type ConnectorsHTTPHandler struct {
	connectorsService ConnectorsService
}

//nolint:interfacebloat // the connectors service contract grows with connector features
type ConnectorsService interface {
	GetConnectors(
		ctx context.Context,
		filter *domain.ConnectorFilter,
	) ([]domain.Connector, error)

	GetConnectorByID(
		ctx context.Context,
		name string,
	) (domain.Connector, error)

	GetConnectorPlugins(
		ctx context.Context,
	) ([]domain.ConnectorPlugin, error)

	GetConnectorPluginSchema(
		ctx context.Context,
		pluginID string,
		filter *domain.ConnectorPluginSchemaFilter,
	) (domain.ConnectorPluginSchema, error)

	GetSMTPlugins(
		ctx context.Context,
	) ([]domain.ConnectorPlugin, error)

	GetSMTPluginSchema(
		ctx context.Context,
		pluginID string,
		filter *domain.ConnectorPluginSchemaFilter,
	) (domain.ConnectorPluginSchema, error)

	CreateConnector(
		ctx context.Context,
		name string,
		config map[string]string,
	) (domain.Connector, error)

	GetTaskByID(
		ctx context.Context,
		name string,
		taskID int,
	) (domain.Task, error)

	RestartTask(
		ctx context.Context,
		name string,
		taskID int,
	) error

	UpdateConnector(
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

	RestartConnector(
		ctx context.Context,
		name string,
		includeTasks bool,
		onlyFailed bool,
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
			Path:    connectorsByIDPath,
			Handler: h.GetConnector,
			Roles:   []string{domain.RoleUser},
		},
		{
			Method:  http.MethodDelete,
			Path:    connectorsByIDPath,
			Handler: h.DeleteConnector,
			Roles:   []string{domain.RoleUser},
		},
		{
			Method:  http.MethodPatch,
			Path:    connectorsByIDPath,
			Handler: h.UpdateConnector,
			Roles:   []string{domain.RoleUser},
		},
		{
			Method:  http.MethodGet,
			Path:    "/connectors/{id}/tasks/{task_id}",
			Handler: h.GetConnectorTask,
			Roles:   []string{domain.RoleUser},
		},
		{
			Method:  http.MethodPost,
			Path:    "/connectors/{id}/tasks/{task_id}/restart",
			Handler: h.RestartConnectorTask,
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
		{
			Method:  http.MethodPost,
			Path:    "/connectors/{id}/restart",
			Handler: h.RestartConnector,
			Roles:   []string{domain.RoleUser},
		},
		{
			Method:  http.MethodGet,
			Path:    "/connector-plugins",
			Handler: h.GetConnectorPlugins,
			Roles:   []string{domain.RoleUser},
		},
		{
			Method:  http.MethodGet,
			Path:    "/connector-plugins/{id}/schema",
			Handler: h.GetConnectorPluginSchema,
			Roles:   []string{domain.RoleUser},
		},
		{
			Method:  http.MethodGet,
			Path:    "/smt-plugins",
			Handler: h.GetSMTPlugins,
			Roles:   []string{domain.RoleUser},
		},
		{
			Method:  http.MethodGet,
			Path:    "/smt-plugins/{id}/schema",
			Handler: h.GetSMTPluginSchema,
			Roles:   []string{domain.RoleUser},
		},
	}
}
