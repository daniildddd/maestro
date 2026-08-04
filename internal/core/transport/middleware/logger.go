package middleware

import (
	"net/http"

	"go.uber.org/zap"

	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
)

func Logger(log *core_logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			requestID := reqctx.RequestId(ctx)

			l := log.With(
				zap.String("request_id", requestID.String()),
				zap.String("url", r.URL.Path),
			)

			ctx = core_logger.ToContext(ctx, l)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
