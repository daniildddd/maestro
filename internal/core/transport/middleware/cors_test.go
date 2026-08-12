package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/transport/middleware"
)

func TestCORS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		allowedOrigins []string
		origin         string
		method         string
		wantStatus     int
		wantCORS       bool
		wantPreflight  bool
	}{
		{
			name:           "allowed origin sets CORS headers",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "http://localhost:3000",
			method:         http.MethodGet,
			wantStatus:     http.StatusOK,
			wantCORS:       true,
		},
		{
			name:           "allowed origin POST passes through",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "http://localhost:3000",
			method:         http.MethodPost,
			wantStatus:     http.StatusOK,
			wantCORS:       true,
		},
		{
			name:           "disallowed origin omits CORS headers",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "http://evil.com",
			method:         http.MethodGet,
			wantStatus:     http.StatusOK,
			wantCORS:       false,
		},
		{
			name:           "missing origin omits CORS headers",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "",
			method:         http.MethodGet,
			wantStatus:     http.StatusOK,
			wantCORS:       false,
		},
		{
			name:           "allowed origin OPTIONS returns 204 with preflight",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "http://localhost:3000",
			method:         http.MethodOptions,
			wantStatus:     http.StatusNoContent,
			wantCORS:       true,
			wantPreflight:  true,
		},
		{
			name:           "disallowed origin OPTIONS returns 204 without preflight",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "http://evil.com",
			method:         http.MethodOptions,
			wantStatus:     http.StatusNoContent,
			wantCORS:       false,
			wantPreflight:  false,
		},
		{
			name:           "missing origin OPTIONS returns 204 without preflight",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "",
			method:         http.MethodOptions,
			wantStatus:     http.StatusNoContent,
			wantCORS:       false,
			wantPreflight:  false,
		},
		{
			name:           "second allowed origin matches",
			allowedOrigins: []string{"http://localhost:3000", "http://localhost:5173"},
			origin:         "http://localhost:5173",
			method:         http.MethodGet,
			wantStatus:     http.StatusOK,
			wantCORS:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			headers := http.Header{}
			if tt.origin != "" {
				headers.Set("Origin", tt.origin)
			}
			req := newTestRequest(t, tt.method, "/", headers)
			rec := httptest.NewRecorder()

			chained := middleware.CORS(tt.allowedOrigins)(nextHandler)
			chained.ServeHTTP(rec, req)

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCORS {
				must.Equal(tt.origin, rec.Header().Get("Access-Control-Allow-Origin"))
				must.Equal("true", rec.Header().Get("Access-Control-Allow-Credentials"))
				must.Equal("Origin", rec.Header().Get("Vary"))
			} else {
				must.Empty(rec.Header().Get("Access-Control-Allow-Origin"))
				must.Empty(rec.Header().Get("Access-Control-Allow-Credentials"))
				must.Empty(rec.Header().Get("Vary"))
			}

			if tt.wantPreflight {
				must.Equal("GET, POST, PATCH, DELETE, OPTIONS", rec.Header().Get("Access-Control-Allow-Methods"))
				must.Equal("Content-Type, Authorization", rec.Header().Get("Access-Control-Allow-Headers"))
				must.Equal("600", rec.Header().Get("Access-Control-Max-Age"))
			} else {
				must.Empty(rec.Header().Get("Access-Control-Allow-Methods"))
				must.Empty(rec.Header().Get("Access-Control-Allow-Headers"))
				must.Empty(rec.Header().Get("Access-Control-Max-Age"))
			}
		})
	}
}
