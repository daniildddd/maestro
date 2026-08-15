package transport

import (
	"context"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/domain"
	core_http_server "github.com/daniildddd/maestro/internal/core/transport/server"
)

const refreshTokenCookieName = "refresh_token"

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
		cfg:         cfg,
	}
}

func (h *AuthHTTPHandler) PrivateRoutes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/logout",
			Handler: h.logout,
		},
	}
}

func (h *AuthHTTPHandler) PublicRoutes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/Login",
			Handler: h.Login,
		},
		{
			Method:  http.MethodPost,
			Path:    "/refresh",
			Handler: h.refresh,
		},
	}
}
