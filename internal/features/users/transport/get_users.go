package transport

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type GetUsersResponse struct {
	Data []UserResponse `json:"data"`
	Meta Meta           `json:"meta"`
}

type UserResponse struct {
	ID        string     `json:"id"`
	Username  string     `json:"username"`
	Role      string     `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type Meta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func (h *UsersHTTPHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	const op = "users.transport.GetUsers"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	q := r.URL.Query()

	page, err := request.GetIntQueryParam(r, "page")
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: get page: %w", op, err),
		)

		return
	}

	limit, err := request.GetIntQueryParam(r, "limit")
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: get limit: %w", op, err),
		)

		return
	}

	filter, err := domain.NewUserFilter(
		page,
		limit,
		q.Get("username"),
		q.Get("role"),
	)
	if err != nil {
		var appErr *errs.AppError

		switch {
		case errors.Is(err, domain.ErrInvalidRole), errors.Is(err, domain.ErrInvalidUsername):
			appErr = errs.ErrValidationFailed
		default:
			appErr = errs.ErrInternal
		}

		responseHandler.ErrorResponse(
			fmt.Errorf("%s: build filter: %w: %v", op, appErr, err),
		)

		return
	}

	users, err := h.usersService.GetUsers(ctx, filter)
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("%s: get users: %w", op, err),
		)

		return
	}

	responseHandler.JSONResponse(
		GetUsersResponse{
			Data: usersResponseFromDomain(users),
			Meta: Meta{
				Page:  filter.Page,
				Limit: filter.Limit,
			},
		},
		http.StatusOK,
	)
}

func usersResponseFromDomain(users []domain.User) []UserResponse {
	resp := make([]UserResponse, 0, len(users))

	for _, u := range users {
		resp = append(resp, userResponseFromDomain(u))
	}

	return resp
}

func userResponseFromDomain(u domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID.String(),
		Username:  u.Username,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
