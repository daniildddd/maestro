package middleware

import (
	"net/http"

	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const requestIDHeader = "X-Request-ID"

func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIDHeader)
			if requestID == "" {
				requestID = uuid.New().String()
			}

			r.Header.Set(requestIDHeader, requestID)
			w.Header().Set(requestIDHeader, requestID)

			next.ServeHTTP(w, r)
		})
	}
}

func Logger(log *core_logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIDHeader)

			l := log.With(
				zap.String("request_id", requestID),
				zap.String("url", r.URL.Path),
			)

			ctx := core_logger.ToContext(r.Context(), l)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				log := core_logger.FromContext(r.Context())
				responseHandler := core_http_response.NewHTTPResponseHandler(w, log)
				if p := recover(); p != nil {
					responseHandler.PanicResponse(
						p,
						"internal error while processing request",
					)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
