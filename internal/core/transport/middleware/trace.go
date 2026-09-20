package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
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
				zap.String("method", logger.Sanitize(r.Method)),
				zap.String("path", logger.Sanitize(r.URL.Path)),
				zap.String("client_ip", logger.Sanitize(reqctx.ClientIP(r.Context()))),
				zap.String("user_agent", logger.Sanitize(reqctx.UserAgent(r.Context()))),
				zap.Duration("latency", time.Since(before)),
			}

			if rw.AuthDone {
				fields = append(fields,
					zap.String("user_id", rw.UserID.String()),
					zap.String("role", rw.Role),
				)
			}

			if !rw.Written() {
				log.Warn("request completed without response", fields...)

				return
			}

			fields = append(fields, zap.Int("status", rw.GetStatusCode()))

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
