package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type GetMeResponse UserDTOResponse

func (h *UsersHTTPHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	const op = "users.transport.GetMe"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	userID := reqctx.UserID(ctx)

	user, err := h.usersService.GetUserByID(ctx, userID)
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: get me: %w", op, err),
		)

		return
	}

	responseHandler.JSONResponse(
		GetMeResponse(userDTOFromDomain(user)),
		http.StatusOK,
	)
}
