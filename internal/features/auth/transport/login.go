package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type LoginUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type LoginUserResponse struct {
	AccessToken string `json:"access_token"`
	Username    string `json:"username"`
}

func (h *AuthHTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "auth.transport.Login"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	var loginRequest LoginUserRequest

	if err := request.DecodeAndValidate(w, r, &loginRequest); err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf(
				"%s: failed decode and validate HTTP request: %w",
				op,
				err,
			),
		)

		return
	}

	loginCredentials, err := h.authService.Login(
		ctx,
		loginRequest.Username,
		loginRequest.Password,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf(
				"%s: Login: %w",
				op,
				err,
			),
		)

		return
	}

	//nolint:gosec // controlled by config; must be true in production
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    loginCredentials.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Domain:   h.cfg.CookieDomain,
		Expires:  loginCredentials.ExpiresAt,
	})

	loginResp := loginResponseFromDomain(
		loginCredentials,
		loginRequest.Username,
	)

	responseHandler.JSONResponse(
		loginResp,
		http.StatusOK,
	)
}

func loginResponseFromDomain(
	credentials domain.TokenPair,
	username string,
) LoginUserResponse {
	return LoginUserResponse{
		AccessToken: credentials.AccessToken,
		Username:    username,
	}
}
