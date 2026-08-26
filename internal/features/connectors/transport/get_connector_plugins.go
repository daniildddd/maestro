package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/logger"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type PluginListItemResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func pluginListItemsFromDomain(plugins []domain.ConnectorPlugin) []PluginListItemResponse {
	resp := make([]PluginListItemResponse, 0, len(plugins))

	for _, p := range plugins {
		resp = append(resp, PluginListItemResponse{
			ID:   p.ID,
			Name: p.Name,
		})
	}

	return resp
}

func (h *ConnectorsHTTPHandler) GetConnectorPlugins(w http.ResponseWriter, r *http.Request) {
	const op = "connectors.transport.GetConnectorPlugins"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	plugins, err := h.connectorsService.GetConnectorPlugins(ctx)
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: list connector plugins: %w", op, err))

		return
	}

	responseHandler.JSONResponse(pluginListItemsFromDomain(plugins), http.StatusOK)
}
