package transport

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type GetLogsResponse struct {
	Data    []AuditEventDTOResponse `json:"data"`
	Meta    PaginationMeta          `json:"meta"`
	HasMore bool                    `json:"has_more"`
}

type PaginationMeta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func (h *AuditHTTPHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	const op = "audit.transport.GetLogs"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	q := r.URL.Query()

	page, err := request.GetIntQueryParam(r, "page")
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: get page: %w", op, err),
		)

		return
	}

	limit, err := request.GetIntQueryParam(r, "limit")
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: get limit: %w", op, err),
		)

		return
	}

	filter, err := domain.NewAuditLogFilter(
		page,
		limit,
		q.Get("action"),
		q.Get("actor"),
	)
	if err != nil {
		var appErr *errs.AppError

		if errors.Is(err, domain.ErrInvalidAuditAction) {
			appErr = errs.ErrValidationFailed
		}

		responseHandler.ErrorResponse(
			fmt.Errorf("%s: build filter: %w: %v", op, appErr, err),
		)

		return
	}

	logs, hasMore, err := h.auditService.GetLogs(ctx, filter)
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: get logs: %w", op, err),
		)

		return
	}

	responseHandler.JSONResponse(
		GetLogsResponse{
			Data:    auditEventsDTOFromDomains(logs),
			Meta:    PaginationMeta{Page: filter.Page, Limit: filter.Limit},
			HasMore: hasMore,
		},
		http.StatusOK,
	)
}
