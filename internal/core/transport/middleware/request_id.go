package middleware

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
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

			ctx = reqctx.WithRequestID(ctx, id)
			w.Header().Set(requestIDHeader, id.String())

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func isValidRequestID(requestID string) (uuid.UUID, bool) {
	v, err := uuid.Parse(requestID)
	if err != nil {
		return uuid.Nil, false
	}

	if v == uuid.Nil {
		return uuid.Nil, false
	}

	return v, true
}
