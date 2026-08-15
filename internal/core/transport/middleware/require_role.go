package middleware

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/daniildddd/maestro/internal/core/errs"
	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func RequireRole(roles ...string) Middleware {
	const op = "transport.middleware.RequireRole"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			role := reqctx.Role(ctx)

			if slices.Contains(roles, role) {
				next.ServeHTTP(w, r)

				return
			}

			log := core_logger.FromContext(ctx)
			responseHandler := core_http_response.NewHTTPResponseHandler(w, log)
			responseHandler.ErrorResponse(
				fmt.Errorf("%s: %w", op, errs.ErrForbidden),
			)
		})
	}
}
