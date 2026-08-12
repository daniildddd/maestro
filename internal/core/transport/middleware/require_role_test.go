package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/middleware"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
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
			roles:         []string{"viewer", "admin"},
			wantForbidden: false,
		},
		{
			name:          "unknown role forbidden",
			role:          "non-accept-role",
			roles:         []string{"viewer", "admin"},
			wantForbidden: true,
		},
		{
			name:          "empty role forbidden",
			role:          "",
			roles:         []string{"viewer", "admin"},
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
			chained := middleware.RequireRole(tt.roles...)(nextHandler)
			chained.ServeHTTP(rec, req)

			if tt.wantForbidden {
				must.Equal(http.StatusForbidden, rec.Code)
				must.Contains(rec.Body.String(), "forbidden")

				return
			}

			must.Equal(http.StatusOK, rec.Code)
		})
	}
}
