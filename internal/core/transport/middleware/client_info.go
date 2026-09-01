package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
)

const (
	clientIPHeader        = "X-Forwarded-For"
	clientUserAgentHeader = "User-Agent"
)

func ClientInfo() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := reqctx.WithClientInfo(r.Context(), clientIP(r), r.Header.Get(clientUserAgentHeader))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get(clientIPHeader); xff != "" {
		if first, _, ok := strings.Cut(xff, ","); ok {
			return strings.TrimSpace(first)
		}

		return strings.TrimSpace(xff)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}
