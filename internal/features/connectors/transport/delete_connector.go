package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func (h *ConnectorsHTTPHandler) DeleteConnector(w http.ResponseWriter, r *http.Request) {
	const op = "connectors.transport.DeleteConnector"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	connectorID, err := request.GetStringPathParam(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: parse path id: %w", op, err))

		return
	}

	if err := h.connectorsService.Delete(ctx, connectorID); err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: delete connector: %w", op, err))

		return
	}

	responseHandler.NoContent()
}
