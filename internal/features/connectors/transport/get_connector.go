package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type ConnectorDetailResponse struct {
	Name       string          `json:"name"`
	PluginType string          `json:"plugin_type"`
	Config     SourceConfigDTO `json:"config"`
	Status     string          `json:"status"`
	WorkerID   string          `json:"worker_id"`
	TasksCount int             `json:"tasks_count"`
	Tasks      []TaskResponse  `json:"tasks"`
}

type SourceConfigDTO struct {
	DatabaseHostname string  `json:"database_hostname,omitempty"`
	DatabasePort     string  `json:"database_port,omitempty"`
	DatabaseUser     string  `json:"database_user,omitempty"`
	DatabaseDBName   *string `json:"database_dbname,omitempty"`
	PluginName       *string `json:"plugin_name,omitempty"`
}

type TaskResponse struct {
	ID       int    `json:"id"`
	State    string `json:"state"`
	WorkerID string `json:"worker_id"`
}

func connectorDetailResponseFromDomain(connector domain.Connector) ConnectorDetailResponse {
	config := SourceConfigDTO{
		DatabaseHostname: connector.Config.Hostname,
		DatabasePort:     connector.Config.Port,
		DatabaseUser:     connector.Config.User,
		DatabaseDBName:   connector.Config.DBName,
		PluginName:       connector.Config.PluginName,
	}

	tasks := make([]TaskResponse, 0, len(connector.Tasks))
	for _, task := range connector.Tasks {
		tasks = append(tasks, TaskResponse{
			ID:       task.ID,
			State:    task.State,
			WorkerID: task.WorkerID,
		})
	}

	return ConnectorDetailResponse{
		Name:       connector.Name,
		PluginType: connector.PluginType,
		Config:     config,
		Status:     connector.Status,
		WorkerID:   connector.WorkerID,
		TasksCount: connector.TasksCount,
		Tasks:      tasks,
	}
}

func (h *ConnectorsHTTPHandler) GetConnector(w http.ResponseWriter, r *http.Request) {
	const op = "connectors.transport.GetConnector"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	name, err := request.GetStringPathParam(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: parse path id: %w", op, err))

		return
	}

	connector, err := h.connectorsService.GetConnectorByID(ctx, name)
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: get connector: %w", op, err))

		return
	}

	responseHandler.JSONResponse(connectorDetailResponseFromDomain(connector), http.StatusOK)
}
