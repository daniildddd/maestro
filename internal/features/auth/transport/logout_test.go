package transport_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/logger"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
	"github.com/daniildddd/maestro/internal/features/auth/transport"
)

func newLogoutRequest(t *testing.T, cookieValue string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/logout", http.NoBody)

	if cookieValue != "" {
		req.AddCookie(&http.Cookie{
			Name:     "refresh_token",
			Value:    cookieValue,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	return req.WithContext(logger.ToContext(req.Context(), nopLogger()))
}

func TestLogout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cookieValue string
		cfg         transport.Config
		setupMock   func(m *MockAuthService)
		wantStatus  int
		wantCode    string
		wantClear   bool
	}{
		{
			name:        "success clears cookie and returns 204",
			cookieValue: "refresh-token",
			cfg: transport.Config{
				CookieSecure: true,
				CookieDomain: "example.com",
			},
			setupMock: func(m *MockAuthService) {
				m.EXPECT().
					Logout(mock.Anything, "refresh-token").
					Return(nil).
					Once()
			},
			wantStatus: http.StatusNoContent,
			wantClear:  true,
		},
		{
			name:       "missing cookie returns 204 without calling service",
			setupMock:  func(_ *MockAuthService) {},
			wantStatus: http.StatusNoContent,
		},
		{
			name:        "service error returns error response without clearing cookie",
			cookieValue: "refresh-token",
			setupMock: func(m *MockAuthService) {
				m.EXPECT().
					Logout(mock.Anything, "refresh-token").
					Return(errs.ErrInternal).
					Once()
			},
			wantStatus: http.StatusInternalServerError,
			wantCode:   "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			authService := NewMockAuthService(t)
			tt.setupMock(authService)

			handler := newTestHandler(authService, tt.cfg)

			rec := httptest.NewRecorder()
			rw := core_http_response.NewResponseWriter(rec)

			handler.Logout(rw, newLogoutRequest(t, tt.cookieValue))

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body errorResponseBody

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body.Code)
			}

			if tt.wantClear {
				cookies := rec.Result().Cookies()
				must.Len(cookies, 1)
				is.Equal("refresh_token", cookies[0].Name)
				is.Empty(cookies[0].Value)
				is.Equal("/", cookies[0].Path)
				is.True(cookies[0].HttpOnly)
				is.Equal(http.SameSiteLaxMode, cookies[0].SameSite)
				is.Equal(tt.cfg.CookieSecure, cookies[0].Secure)
				is.Equal(tt.cfg.CookieDomain, cookies[0].Domain)
				is.Contains(rec.Header().Get("Set-Cookie"), "Max-Age=0")

				return
			}

			is.Empty(rec.Header().Get("Set-Cookie"))
		})
	}
}
