package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap/zapcore"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/middleware"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
)

func TestLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "logs request_id and url for GET /api/v1/users",
			method: http.MethodGet,
			path:   "/api/v1/users",
		},
		{
			name:   "logs request_id and url for POST /api/v1/login",
			method: http.MethodPost,
			path:   "/api/v1/login",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)

			requestID := uuid.New()
			log, recordedLogs := newObservableLogger(t, zapcore.InfoLevel)

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				logger.FromContext(r.Context()).Info("test")
				w.WriteHeader(http.StatusOK)
			})

			req := newTestRequest(t, tt.method, tt.path, nil)
			req = req.WithContext(reqctx.WithRequestID(req.Context(), requestID))
			rec := httptest.NewRecorder()

			middleware.Logger(log)(nextHandler).ServeHTTP(rec, req)

			must.Equal(http.StatusOK, rec.Code)

			logs := recordedLogs.All()
			must.Len(logs, 1)
			must.Equal(zapcore.InfoLevel, logs[0].Level)
			must.Contains(logs[0].Message, "test")

			contextMap := logs[0].ContextMap()
			must.Len(contextMap, 2)
			must.Equal(requestID.String(), contextMap["request_id"])
			must.Equal(tt.path, contextMap["url"])
		})
	}

	t.Run("panics without request_id in context", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		log, _ := newObservableLogger(t, zapcore.InfoLevel)

		nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			t.Fatal("next handler must not be called")
		})

		req := newTestRequest(t, http.MethodGet, "/api/v1/users", nil)
		rec := httptest.NewRecorder()

		must.Panics(func() {
			middleware.Logger(log)(nextHandler).ServeHTTP(rec, req)
		})
	})
}
