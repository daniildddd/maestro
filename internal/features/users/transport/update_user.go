package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type UpdateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
}

func (h *UsersHTTPHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	const op = "users.transport.UpdateUser"

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

	var req UpdateUserRequest

	if err := request.DecodeAndValidate(w, r, &req); err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: decode request: %w", op, err),
		)

		return
	}

	user, err := h.usersService.UpdateUser(ctx, id, req.Username)
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: update user: %w", op, err),
		)

		return
	}

	responseHandler.JSONResponse(
		userResponseFromDomain(user),
		http.StatusOK,
	)
}
