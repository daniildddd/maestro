package transport

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	core_http_server "github.com/daniildddd/maestro/internal/core/transport/server"
)

type UsersHTTPHandler struct {
	usersService UsersService
}

type UsersService interface {
	GetUsers(
		ctx context.Context,
		filter domain.UserFilter,
	) ([]domain.User, error)

	CreateUser(
		ctx context.Context,
		username string,
		password string,
		role string,
	) (domain.User, error)

	GetUserByID(
		ctx context.Context,
		id uuid.UUID,
	) (domain.User, error)

	DeleteUser(
		ctx context.Context,
		id uuid.UUID,
	) error

	DeleteMe(
		ctx context.Context,
		userID uuid.UUID,
		password string,
	) error

	ChangePassword(
		ctx context.Context,
		id uuid.UUID,
		newPassword string,
	) error
}

func NewUsersHTTPHandler(
	usersService UsersService,
) *UsersHTTPHandler {
	return &UsersHTTPHandler{
		usersService: usersService,
	}
}

func (h *UsersHTTPHandler) PrivateRoutes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/users",
			Handler: h.CreateUser,
			Roles:   []string{domain.RoleAdmin},
		},
		{
			Method:  http.MethodGet,
			Path:    "/users",
			Handler: h.GetUsers,
			Roles:   []string{domain.RoleAdmin},
		},
		{
			Method:  http.MethodPatch,
			Path:    "/users/{id}/password",
			Handler: h.ChangePassword,
			Roles:   []string{domain.RoleAdmin},
		},
		{
			Method:  http.MethodGet,
			Path:    "/users/{id}",
			Handler: h.GetUserByID,
			Roles:   []string{domain.RoleUser, domain.RoleAdmin},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/users/{id}",
			Handler: h.DeleteUser,
			Roles:   []string{domain.RoleAdmin},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/users/me",
			Handler: h.DeleteMe,
			Roles:   []string{domain.RoleUser, domain.RoleAdmin},
		},
	}
}
