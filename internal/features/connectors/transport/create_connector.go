package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type ConnectorCreateRequest struct {
	Name       string            `json:"name"        validate:"required,min=1,max=128"`
	PluginType string            `json:"plugin_type" validate:"required"`
	Config     map[string]string `json:"config"      validate:"required"`
}

type ConnectorCreateResponse struct {
	Name       string          `json:"name"`
	PluginType string          `json:"plugin_type"`
	Status     string          `json:"status"`
	Config     SourceConfigDTO `json:"config"`
	TasksCount int             `json:"tasks_count"`
	Tasks      []TaskResponse  `json:"tasks"`
}

func connectorCreateResponseFromDomain(connector domain.Connector) ConnectorCreateResponse {
	return ConnectorCreateResponse{
		Name:       connector.Name,
		PluginType: connector.PluginType,
		Status:     connector.Status,
		Config: SourceConfigDTO{
			DatabaseHostname: connector.Config.Hostname,
			DatabasePort:     connector.Config.Port,
			DatabaseUser:     connector.Config.User,
			DatabaseDBName:   connector.Config.DBName,
			PluginName:       connector.Config.PluginName,
		},
		TasksCount: connector.TasksCount,
		Tasks:      []TaskResponse{},
	}
}

func (h *ConnectorsHTTPHandler) CreateConnector(w http.ResponseWriter, r *http.Request) {
	const op = "connectors.transport.CreateConnector"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	var req ConnectorCreateRequest

	if err := request.DecodeAndValidate(w, r, &req); err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: decode request: %w", op, err))

		return
	}

	connector, err := h.connectorsService.CreateConnector(ctx, req.Name, req.Config)
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: create connector: %w", op, err))

		return
	}

	responseHandler.JSONResponse(connectorCreateResponseFromDomain(connector), http.StatusCreated)
}
