package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func (h *ConnectorsHTTPHandler) GetSMTPluginSchema(w http.ResponseWriter, r *http.Request) {
	const op = "connectors.transport.GetSMTPluginSchema"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	pluginID, err := request.GetStringPathParam(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: parse path id: %w", op, err))

		return
	}

	filter, err := domain.NewConnectorPluginSchemaFilter(r.URL.Query().Get("importance"))
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: build filter: %w: %v", op, errs.ErrInvalidQueryParam, err),
		)

		return
	}

	schema, err := h.connectorsService.GetSMTPluginSchema(ctx, pluginID, filter)
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: get smt plugin schema: %w", op, err))

		return
	}

	responseHandler.JSONResponse(pluginSchemaResponseFromDomain(schema), http.StatusOK)
}
