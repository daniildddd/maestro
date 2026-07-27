package middleware

import (
	"net/http"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const requestIDHeader = "X-Request-ID"

func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := logger.FromContext(ctx)

			received := r.Header.Get(requestIDHeader)
			id, ok := isValidRequestID(received)
			if !ok {
				log.Debug(
					"invalid or missing X-Request-ID, generating new one",
					zap.String("received", received),
				)
				id = uuid.New()
			}

			ctx = reqctx.WithRequestId(ctx, id)
			w.Header().Set(requestIDHeader, id.String())

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func isValidRequestID(requestId string) (uuid.UUID, bool) {
	v, err := uuid.Parse(requestId)
	if err != nil {
		return uuid.Nil, false
	}

	return v, true
}
