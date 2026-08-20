package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type GetUserByIDResponse UserDTOResponse

func (h *UsersHTTPHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	const op = "users.transport.GetUserByID"

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

	user, err := h.usersService.GetUserByID(ctx, id)
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: get user by id: %w", op, err),
		)

		return
	}

	responseHandler.JSONResponse(
		GetUserByIDResponse(userDTOFromDomain(user)),
		http.StatusOK,
	)
}
