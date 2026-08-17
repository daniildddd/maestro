package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func (h *UsersHTTPHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	const op = "users.transport.DeleteUser"

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

	err = h.usersService.DeleteUser(ctx, id)
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: delete user: %w", op, err),
		)

		return
	}

	responseHandler.NoContent()
}
