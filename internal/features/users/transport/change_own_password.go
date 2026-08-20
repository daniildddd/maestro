package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type ChangeOwnPasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required,min=8,max=128"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

func (h *UsersHTTPHandler) ChangeOwnPassword(w http.ResponseWriter, r *http.Request) {
	const op = "users.transport.ChangeOwnPassword"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	userID := reqctx.UserID(ctx)

	var req ChangeOwnPasswordRequest

	if err := request.DecodeAndValidate(w, r, &req); err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: decode request (user_id=%s): %w", op, userID, err),
		)

		return
	}

	err := h.usersService.ChangeOwnPassword(ctx, userID, req.OldPassword, req.NewPassword)
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: change own password: %w", op, err),
		)

		return
	}

	responseHandler.NoContent()
}
