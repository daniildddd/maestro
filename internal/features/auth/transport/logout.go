package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/response"
)

func (h *AuthHTTPHandler) Logout(w http.ResponseWriter, r *http.Request) {
	const op = "auth.transport.logout"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewHTTPResponseHandler(w, log)

	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie == nil {
		responseHandler.NoContent()

		return
	}

	err = h.authService.Logout(ctx, cookie.Value)
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf(
				"%s: logout: %w",
				op,
				err,
			),
		)

		return
	}

	//nolint:gosec // controlled by config; must be true in production
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Domain:   h.cfg.CookieDomain,
		MaxAge:   -1,
	})

	responseHandler.NoContent()
}
