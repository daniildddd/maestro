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

type PluginSchemaFieldResponse struct {
	Name        string   `json:"name"`
	Label       *string  `json:"label,omitempty"`
	Description *string  `json:"description,omitempty"`
	Type        string   `json:"type"`
	Importance  string   `json:"importance"`
	Required    bool     `json:"required"`
	Default     *string  `json:"default,omitempty"`
	Values      []string `json:"values,omitempty"`
}

type PluginSchemaResponse struct {
	Fields []PluginSchemaFieldResponse `json:"fields"`
}

func pluginSchemaResponseFromDomain(schema domain.ConnectorPluginSchema) PluginSchemaResponse {
	fields := make([]PluginSchemaFieldResponse, 0, len(schema.Fields))

	for _, f := range schema.Fields {
		fields = append(fields, PluginSchemaFieldResponse{
			Name:        f.Name,
			Label:       f.Label,
			Description: f.Description,
			Type:        f.Type,
			Importance:  f.Importance,
			Required:    f.Required,
			Default:     f.Default,
			Values:      f.Values,
		})
	}

	return PluginSchemaResponse{Fields: fields}
}

func (h *ConnectorsHTTPHandler) GetConnectorPluginSchema(w http.ResponseWriter, r *http.Request) {
	const op = "connectors.transport.GetConnectorPluginSchema"

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

	schema, err := h.connectorsService.GetConnectorPluginSchema(ctx, pluginID, filter)
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: get connector plugin schema: %w", op, err))

		return
	}

	responseHandler.JSONResponse(pluginSchemaResponseFromDomain(schema), http.StatusOK)
}
