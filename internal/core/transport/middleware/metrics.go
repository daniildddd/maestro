package middleware

import (
	"net/http"
	"time"

	core_metrics "github.com/daniildddd/maestro/internal/core/metrics"
	"github.com/daniildddd/maestro/internal/core/transport/response"
)

func Metrics(m *core_metrics.Metrics) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)

				return
			}

			rw := response.NewResponseWriter(w)
			start := time.Now()

			next.ServeHTTP(rw, r)

			route := r.Pattern
			if route == "" || route == "/api/v1/{rest...}" {
				route = "unmatched"
			}

			m.Observe(r.Method, route, rw.GetStatusCode(), time.Since(start))
		})
	}
}
