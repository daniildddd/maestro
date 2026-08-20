package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type ChangePasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

func (h *UsersHTTPHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	const op = "users.transport.ChangePassword"

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

	var req ChangePasswordRequest

	if err := request.DecodeAndValidate(w, r, &req); err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: decode request: %w", op, err),
		)

		return
	}

	err = h.usersService.ChangePassword(ctx, id, req.NewPassword)
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: change password: %w", op, err),
		)

		return
	}

	responseHandler.NoContent()
}
