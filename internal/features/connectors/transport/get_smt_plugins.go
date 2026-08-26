package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func (h *ConnectorsHTTPHandler) GetSMTPlugins(w http.ResponseWriter, r *http.Request) {
	const op = "connectors.transport.GetSMTPlugins"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	plugins, err := h.connectorsService.GetSMTPlugins(ctx)
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: list smt plugins: %w", op, err))

		return
	}

	responseHandler.JSONResponse(pluginListItemsFromDomain(plugins), http.StatusOK)
}
