package transport

import (
	"context"

	"github.com/daniildddd/maestro/internal/core/domain"
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
