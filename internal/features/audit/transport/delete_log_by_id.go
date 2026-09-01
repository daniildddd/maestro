package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func (h *AuditHTTPHandler) DeleteLogByID(w http.ResponseWriter, r *http.Request) {
	const op = "audit.transport.DeleteLogByID"

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

	if err = h.auditService.DeleteLogByID(ctx, id); err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: delete log by id: %w", op, err),
		)

		return
	}

	responseHandler.NoContent()
}
