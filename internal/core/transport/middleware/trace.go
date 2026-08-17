package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/daniildddd/maestro/internal/core/logger"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func Trace() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := logger.FromContext(ctx)
			rw := core_http_response.NewResponseWriter(w)

			before := time.Now()

			next.ServeHTTP(rw, r)

			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", rw.GetStatusCode()),
				zap.Duration("latency", time.Since(before)),
			}

			appErr := rw.AppErr
			if appErr != nil {
				fields = append(fields,
					zap.String("code", appErr.Code),
					zap.Error(rw.RawErr),
				)

				if ce := log.Check(appErr.LogLevel, "request completed with error"); ce != nil {
					ce.Write(fields...)
				}

				return
			}

			log.Info("request completed successfully", fields...)
		})
	}
}
