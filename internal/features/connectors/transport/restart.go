package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func (h *ConnectorsHTTPHandler) RestartConnector(w http.ResponseWriter, r *http.Request) {
	const op = "connectors.transport.RestartConnector"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	connectorID, err := request.GetStringPathParam(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: parse path id: %w", op, err))

		return
	}

	includeTasks, err := request.GetBoolQueryParam(r, "include_tasks")
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: get include_tasks: %w", op, err))

		return
	}

	onlyFailed, err := request.GetBoolQueryParam(r, "only_failed")
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: get only_failed: %w", op, err))

		return
	}

	connector, err := h.connectorsService.RestartConnector(ctx, connectorID, includeTasks, onlyFailed)
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: restart connector: %w", op, err))

		return
	}

	responseHandler.JSONResponse(connectorDetailResponseFromDomain(connector), http.StatusOK)
}
