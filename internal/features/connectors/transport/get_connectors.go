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

func (h *ConnectorsHTTPHandler) GetConnectors(w http.ResponseWriter, r *http.Request) {
	const op = "connectors.transport.GetConnectors"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	q := r.URL.Query()

	page, err := request.GetIntQueryParam(r, "page")
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: get page: %w", op, err))

		return
	}

	limit, err := request.GetIntQueryParam(r, "limit")
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: get limit: %w", op, err))

		return
	}

	filter, err := domain.NewConnectorFilter(
		page,
		limit,
		q.Get("status"),
		q.Get("search"),
	)
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: build filter: %w: %v", op, errs.ErrInvalidQueryParam, err),
		)

		return
	}

	connectors, err := h.connectorsService.List(ctx, filter)
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: list connectors: %w", op, err))

		return
	}

	responseHandler.JSONResponse(
		ConnectorListResponse{
			Data: connectorsResponseFromDomain(connectors),
			Meta: Meta{Page: filter.Page, Limit: filter.Limit},
		},
		http.StatusOK,
	)
}

type ConnectorListResponse struct {
	Data []ConnectorResponse `json:"data"`
	Meta Meta                `json:"meta"`
}

type ConnectorResponse struct {
	Name       string `json:"name"`
	PluginType string `json:"plugin_type"`
	Status     string `json:"status"`
	TasksCount int    `json:"tasks_count"`
}

type Meta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func connectorsResponseFromDomain(connectors []domain.Connector) []ConnectorResponse {
	resp := make([]ConnectorResponse, 0, len(connectors))

	for _, c := range connectors {
		resp = append(resp, connectorResponseFromDomain(c))
	}

	return resp
}

func connectorResponseFromDomain(c domain.Connector) ConnectorResponse {
	return ConnectorResponse{
		Name:       c.Name,
		PluginType: c.PluginType,
		Status:     c.Status,
		TasksCount: c.TasksCount,
	}
}
