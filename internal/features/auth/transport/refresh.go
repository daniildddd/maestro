package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/logger"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type refreshResponse struct {
	AccessToken string `json:"access_token"`
}

func (h *AuthHTTPHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	const op = "auth.transport.refresh"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf(
				"%s: missing refresh token cookie: %w",
				op,
				errs.ErrMissingRefreshToken,
			),
		)

		return
	}

	refreshToken := cookie.Value

	tokenPair, err := h.authService.Refresh(ctx, refreshToken)
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf(
				"%s: refresh tokens: %w",
				op,
				err,
			),
		)

		return
	}

	//nolint:gosec // controlled by config; must be true in production
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    tokenPair.RefreshToken,
		Path:     "/",
		Domain:   h.cfg.CookieDomain,
		Expires:  tokenPair.ExpiresAt,
		Secure:   h.cfg.CookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	refreshResp := refreshResponse{
		AccessToken: tokenPair.AccessToken,
	}

	responseHandler.JSONResponse(
		refreshResp,
		http.StatusOK,
	)
}
