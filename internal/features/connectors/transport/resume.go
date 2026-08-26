package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func (h *ConnectorsHTTPHandler) ResumeConnector(w http.ResponseWriter, r *http.Request) {
	const op = "connectors.transport.ResumeConnector"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	connectorID, err := request.GetStringPathParam(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: parse path id: %w", op, err))

		return
	}

	connector, err := h.connectorsService.ResumeConnector(ctx, connectorID)
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: resume connector: %w", op, err))

		return
	}

	responseHandler.JSONResponse(connectorDetailResponseFromDomain(connector), http.StatusOK)
}
