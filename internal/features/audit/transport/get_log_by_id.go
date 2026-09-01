package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type GetLogByIDResponse AuditEventDTOResponse

func (h *AuditHTTPHandler) GetLogByID(w http.ResponseWriter, r *http.Request) {
	const op = "audit.transport.GetLogByID"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	id, err := request.GetUUIDPathParam(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: parse path id: %w", op, err),
		)

		return
	}

	event, err := h.auditService.GetLogByID(ctx, id)
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: get log by id: %w", op, err),
		)

		return
	}

	responseHandler.JSONResponse(
		GetLogByIDResponse(auditEventDTOFromDomain(event)),
		http.StatusOK,
	)
}
