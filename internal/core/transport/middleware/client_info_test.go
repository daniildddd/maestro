package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/transport/middleware"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
)

func TestClientInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		remoteAddr string
		xffHeader  string
		userAgent  string
		wantIP     string
		wantUA     string
	}{
		{
			name:       "remote addr host used without forwarded header",
			remoteAddr: "192.168.1.10:52341",
			userAgent:  "curl/8.7.1",
			wantIP:     "192.168.1.10",
			wantUA:     "curl/8.7.1",
		},
		{
			name:       "single forwarded value used instead of remote addr",
			remoteAddr: "192.168.1.10:52341",
			xffHeader:  "203.0.113.50",
			userAgent:  "curl/8.7.1",
			wantIP:     "203.0.113.50",
			wantUA:     "curl/8.7.1",
		},
		{
			name:       "first entry of forwarded chain is taken",
			remoteAddr: "192.168.1.10:52341",
			xffHeader:  "203.0.113.50, 10.0.0.1, 10.0.0.2",
			userAgent:  "Mozilla/5.0",
			wantIP:     "203.0.113.50",
			wantUA:     "Mozilla/5.0",
		},
		{
			name:       "forwarded value is trimmed",
			remoteAddr: "192.168.1.10:52341",
			xffHeader:  "  203.0.113.50  ",
			userAgent:  "curl/8.7.1",
			wantIP:     "203.0.113.50",
			wantUA:     "curl/8.7.1",
		},
		{
			name:       "remote addr without port returned as is",
			remoteAddr: "192.168.1.10",
			xffHeader:  "",
			userAgent:  "curl/8.7.1",
			wantIP:     "192.168.1.10",
			wantUA:     "curl/8.7.1",
		},
		{
			name:       "empty user agent passes through",
			remoteAddr: "192.168.1.10:52341",
			xffHeader:  "",
			userAgent:  "",
			wantIP:     "192.168.1.10",
			wantUA:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			must := require.New(t)

			var (
				gotIP string
				gotUA string
			)

			handler := middleware.ClientInfo()(http.HandlerFunc(
				func(_ http.ResponseWriter, r *http.Request) {
					gotIP = reqctx.ClientIP(r.Context())
					gotUA = reqctx.UserAgent(r.Context())
				},
			))

			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			req.RemoteAddr = tt.remoteAddr

			if tt.xffHeader != "" {
				req.Header.Set("X-Forwarded-For", tt.xffHeader)
			}

			if tt.userAgent != "" {
				req.Header.Set("User-Agent", tt.userAgent)
			}

			rec := httptest.NewRecorder()

			must.NotPanics(func() {
				handler.ServeHTTP(rec, req)
			})

			is.Equal(tt.wantIP, gotIP)
			is.Equal(tt.wantUA, gotUA)
		})
	}
}
