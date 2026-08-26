package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type ConnectorUpdateRequest struct {
	Config map[string]string `json:"config" validate:"required"`
}

func (h *ConnectorsHTTPHandler) UpdateConnector(w http.ResponseWriter, r *http.Request) {
	const op = "connectors.transport.UpdateConnector"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	connectorID, err := request.GetStringPathParam(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: parse path id: %w", op, err))

		return
	}

	var req ConnectorUpdateRequest

	if err := request.DecodeAndValidate(w, r, &req); err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: decode request: %w", op, err))

		return
	}

	connector, err := h.connectorsService.UpdateConnector(ctx, connectorID, req.Config)
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: update connector: %w", op, err))

		return
	}

	responseHandler.JSONResponse(connectorDetailResponseFromDomain(connector), http.StatusOK)
}
