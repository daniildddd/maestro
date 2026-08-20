package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type DeleteMeRequest struct {
	Password string `json:"password" validate:"required,min=8,max=128"`
}

func (h *UsersHTTPHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	const op = "users.transport.DeleteMe"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	userID := reqctx.UserID(ctx)

	var deleteMeRequest DeleteMeRequest

	if err := request.DecodeAndValidate(w, r, &deleteMeRequest); err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: decode and validate (user_id=%s): %w", op, userID, err),
		)

		return
	}

	err := h.usersService.DeleteMe(
		ctx,
		userID,
		deleteMeRequest.Password,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: delete me: %w", op, err),
		)

		return
	}

	responseHandler.NoContent()
}
