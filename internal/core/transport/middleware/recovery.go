package middleware

import (
	"net/http"

	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				log := core_logger.FromContext(r.Context())
				responseHandler := core_http_response.NewHTTPResponseHandler(w, log)
				if p := recover(); p != nil {
					responseHandler.PanicResponse(p)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
