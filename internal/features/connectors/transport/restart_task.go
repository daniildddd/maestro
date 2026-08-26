package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func (h *ConnectorsHTTPHandler) RestartConnectorTask(w http.ResponseWriter, r *http.Request) {
	const op = "connectors.transport.RestartConnectorTask"

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

	if err := h.connectorsService.RestartTask(ctx, name, taskID); err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: restart task: %w", op, err))

		return
	}

	responseHandler.NoContent()
}
