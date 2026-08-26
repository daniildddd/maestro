package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type TaskDetailResponse struct {
	ID       int     `json:"id"`
	State    string  `json:"state"`
	WorkerID string  `json:"worker_id"`
	Trace    *string `json:"trace"`
}

func taskDetailResponseFromDomain(task domain.Task) TaskDetailResponse {
	return TaskDetailResponse{
		ID:       task.ID,
		State:    task.State,
		WorkerID: task.WorkerID,
		Trace:    task.Trace,
	}
}

func (h *ConnectorsHTTPHandler) GetConnectorTask(w http.ResponseWriter, r *http.Request) {
	const op = "connectors.transport.GetConnectorTask"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	name, err := request.GetStringPathParam(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: parse path id: %w", op, err))

		return
	}

	taskID, err := request.GetIntPathParam(r, "task_id")
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: parse path task_id: %w", op, err))

		return
	}

	task, err := h.connectorsService.GetTaskByID(ctx, name, taskID)
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: get task: %w", op, err))

		return
	}

	responseHandler.JSONResponse(taskDetailResponseFromDomain(task), http.StatusOK)
}
