package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/daniildddd/maestro/internal/core/errs"
	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/security/access"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
	"go.uber.org/zap"
)

type TokenVerifier interface {
	Verify(
		tokenStr string,
	) (access.AuthUser, error)
}

func Auth(tv TokenVerifier) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := core_logger.FromContext(ctx)

			token, ok := stripBearer(r.Header.Get("Authorization"))
			if !ok {
				writeUnauthorized(ctx, w, "missing or invalid Authorization header")
				return
			}

			claims, err := tv.Verify(token)
			if err != nil {
				log.Warn("verify jwt token", zap.Error(err))
				writeUnauthorized(ctx, w, "invalid or expired token")
				return
			}

			l := log.With(
				zap.String("user_id", claims.UserId.String()),
				zap.String("role", claims.Role),
			)

			ctx = reqctx.WithUserId(ctx, claims.UserId)
			ctx = reqctx.WithRole(ctx, claims.Role)

			ctx = core_logger.ToContext(ctx, l)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func stripBearer(h string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return "", false
	}

	return strings.TrimPrefix(h, prefix), true
}

func writeUnauthorized(ctx context.Context, w http.ResponseWriter, msg string) {
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)
	responseHandler.ErrorResponse(errs.ErrInvalidCredentials, msg)
}
