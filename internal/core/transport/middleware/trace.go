package middleware

import (
	"net/http"
	"time"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
	"github.com/daniildddd/maestro/internal/core/transport/response"
	"go.uber.org/zap"
)

func Trace() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := logger.FromContext(ctx)
			rw := response.NewResponseWriter(w)

			before := time.Now()
			id := reqctx.RequestId(ctx)
			log.Debug(
				">>> incoming HTTP request",
				zap.String("http_method", r.Method),
				zap.String("request_id", id.String()),
				zap.Time("time", before.UTC()),
			)

			next.ServeHTTP(rw, r)

			log.Debug(
				"<<< done HTTP request",
				zap.Int("status_code", rw.GetStatusCode()),
				zap.Duration("latency", time.Since(before)),
				zap.String("request_id", id.String()),
			)
		})
	}
}
