package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Password string `json:"password" validate:"required,min=8,max=128"`
	Role     string `json:"role"     validate:"required"`
}

type CreateUserResponse UserDTOResponse

func (h *UsersHTTPHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	const op = "users.transport.CreateUser"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	var req CreateUserRequest

	if err := request.DecodeAndValidate(w, r, &req); err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: decode request: %w", op, err),
		)

		return
	}

	user, err := h.usersService.CreateUser(
		ctx,
		req.Username,
		req.Password,
		req.Role,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: create user: %w", op, err),
		)

		return
	}

	responseHandler.JSONResponse(
		CreateUserResponse(userDTOFromDomain(user)),
		http.StatusCreated,
	)
}
