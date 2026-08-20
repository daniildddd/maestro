package transport

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type GetUsersResponse struct {
	Data []UserDTOResponse `json:"data"`
	Meta PaginationMeta    `json:"meta"`
}

type PaginationMeta struct {
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
		case errors.Is(err, domain.ErrInvalidRole),
			errors.Is(err, domain.ErrInvalidUsername):
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
			Data: usersDTOFromDomains(users),
			Meta: PaginationMeta{
				Page:  filter.Page,
				Limit: filter.Limit,
			},
		},
		http.StatusOK,
	)
}
