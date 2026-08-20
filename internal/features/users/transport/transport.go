package transport

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	core_http_server "github.com/daniildddd/maestro/internal/core/transport/server"
)

const usersByIDPath = "/users/{id}"

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

	ChangeOwnPassword(
		ctx context.Context,
		userID uuid.UUID,
		oldPassword string,
		newPassword string,
	) error

	UpdateUser(
		ctx context.Context,
		id uuid.UUID,
		username string,
	) (domain.User, error)
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
			Method:  http.MethodPatch,
			Path:    usersByIDPath,
			Handler: h.UpdateUser,
			Roles:   []string{domain.RoleAdmin},
		},
		{
			Method:  http.MethodGet,
			Path:    usersByIDPath,
			Handler: h.GetUserByID,
			Roles:   []string{domain.RoleUser, domain.RoleAdmin},
		},
		{
			Method:  http.MethodDelete,
			Path:    usersByIDPath,
			Handler: h.DeleteUser,
			Roles:   []string{domain.RoleAdmin},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/users/me",
			Handler: h.DeleteMe,
			Roles:   []string{domain.RoleUser, domain.RoleAdmin},
		},
		{
			Method:  http.MethodPatch,
			Path:    "/users/me/password",
			Handler: h.ChangeOwnPassword,
			Roles:   []string{domain.RoleUser, domain.RoleAdmin},
		},
	}
}
