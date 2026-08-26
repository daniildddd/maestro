package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type RestartRequest struct {
	IncludeTasks *bool `json:"include_tasks,omitempty"`
	OnlyFailed   *bool `json:"only_failed,omitempty"`
}

func (h *ConnectorsHTTPHandler) RestartConnector(w http.ResponseWriter, r *http.Request) {
	const op = "connectors.transport.RestartConnector"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	name, err := request.GetStringPathParam(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: parse path id: %w", op, err))

		return
	}

	var req RestartRequest

	if r.ContentLength != 0 {
		if err := request.DecodeAndValidate(w, r, &req); err != nil {
			responseHandler.ErrorResponse(fmt.Errorf("%s: decode request: %w", op, err))

			return
		}
	}

	includeTasks := req.IncludeTasks != nil && *req.IncludeTasks
	onlyFailed := req.OnlyFailed != nil && *req.OnlyFailed

	connector, err := h.connectorsService.RestartConnector(ctx, name, includeTasks, onlyFailed)
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: restart connector: %w", op, err))

		return
	}

	responseHandler.JSONResponse(connectorDetailResponseFromDomain(connector), http.StatusOK)
}
