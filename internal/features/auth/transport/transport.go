package transport

import (
	"context"

	"github.com/daniildddd/maestro/internal/core/domain"
	core_http_server "github.com/daniildddd/maestro/internal/core/transport/server"
)

type AuthHTTPHandler struct {
	authService AuthService
	cfg         Config
}

type AuthService interface {
	Login(
		ctx context.Context,
		username string,
		password string,
	) (domain.TokenPair, error)

	Logout(
		ctx context.Context,
		refreshToken string,
	) error

	Refresh(
		ctx context.Context,
		refreshToken string,
	) (domain.TokenPair, error)
}

func NewAuthHTTPHandler(
	authService AuthService,
	cfg Config,
) *AuthHTTPHandler {
	return &AuthHTTPHandler{
		authService: authService,
		cfg:          cfg,
	}
}

func (h *AuthHTTPHandler) PublicRoutes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  "POST",
			Path:    "/login",
			Handler: h.login,
		},
	}
}