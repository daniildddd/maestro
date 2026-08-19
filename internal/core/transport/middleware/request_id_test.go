package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/middleware"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
)

func TestRequestID(t *testing.T) {
	t.Parallel()

	validID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name             string
		headerValue      string
		expectedGenerate bool
	}{
		{
			name:             "valid id passes through",
			headerValue:      validID,
			expectedGenerate: false,
		},
		{
			name:             "invalid id generates new",
			headerValue:      "not-a-uuid-string",
			expectedGenerate: true,
		},
		{
			name:             "missing id generates new",
			headerValue:      "",
			expectedGenerate: true,
		},
		{
			name:             "nil uuid regenerates",
			headerValue:      "00000000000000000000000000000000",
			expectedGenerate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)

			var capturedID uuid.UUID

			nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				capturedID = reqctx.RequestID(r.Context())
			})

			headers := http.Header{"X-Request-ID": []string{tt.headerValue}}
			req := newTestRequest(t, http.MethodGet, "/", headers)
			req = req.WithContext(core_logger.ToContext(req.Context(), nopLogger()))
			rec := httptest.NewRecorder()

			chained := middleware.RequestID()(nextHandler)
			chained.ServeHTTP(rec, req)

			requestHeader := rec.Header().Get("X-Request-ID")

			must.NotEmpty(requestHeader)
			must.Equal(capturedID.String(), requestHeader)

			parsed, err := uuid.Parse(requestHeader)
			must.NoError(err)
			must.NotEqual(uuid.Nil, parsed)

			if tt.expectedGenerate {
				must.NotEqual(tt.headerValue, requestHeader)

				return
			}

			must.Equal(tt.headerValue, requestHeader)
		})
	}
}
