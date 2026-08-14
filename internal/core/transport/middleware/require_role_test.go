package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/middleware"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func TestRequireRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		role          string
		roles         []string
		wantForbidden bool
	}{
		{
			name:          "allowed role passes through",
			role:          "admin",
			roles:         []string{"user", "admin"},
			wantForbidden: false,
		},
		{
			name:          "unknown role forbidden",
			role:          "non-accept-role",
			roles:         []string{"user", "admin"},
			wantForbidden: true,
		},
		{
			name:          "empty role forbidden",
			role:          "",
			roles:         []string{"user", "admin"},
			wantForbidden: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			rec := httptest.NewRecorder()
			req := newTestRequest(t, http.MethodGet, "/", nil)
			ctx := logger.ToContext(req.Context(), nopLogger())
			ctx = reqctx.WithRole(ctx, tt.role)
			req = req.WithContext(ctx)
			rw := core_http_response.NewResponseWriter(rec)
			chained := middleware.RequireRole(tt.roles...)(nextHandler)
			chained.ServeHTTP(rw, req)

			if tt.wantForbidden {
				must.Equal(http.StatusForbidden, rec.Code)

				var body core_http_response.ErrorResponse
				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				must.Equal("FORBIDDEN", body.Code)
				must.Equal("Forbidden", body.Message)

				return
			}

			must.Equal(http.StatusOK, rec.Code)
		})
	}
}
