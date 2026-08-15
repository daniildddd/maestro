package transport_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/logger"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
	"github.com/daniildddd/maestro/internal/features/auth/transport"
)

func newRefreshRequest(t *testing.T, cookieValue string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/refresh", http.NoBody)

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

func TestRefresh(t *testing.T) {
	t.Parallel()

	expiresAt := time.Now().Add(time.Hour).Truncate(time.Second)

	tests := []struct {
		name        string
		cookieValue string
		cfg         transport.Config
		setupMock   func(m *MockAuthService)
		wantStatus  int
		wantCode    string
		wantToken   string
	}{
		{
			name:        "valid refresh token returns new tokens and rotates cookie",
			cookieValue: "refresh-token",
			cfg: transport.Config{
				CookieSecure: true,
				CookieDomain: "example.com",
			},
			setupMock: func(m *MockAuthService) {
				m.EXPECT().
					Refresh(mock.Anything, "refresh-token").
					Return(domain.TokenPair{
						AccessToken:  "new-access-token",
						RefreshToken: "new-refresh-token",
						ExpiresAt:    expiresAt,
					}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantToken:  "new-access-token",
		},
		{
			name:       "missing cookie returns INVALID_REFRESH_TOKEN",
			setupMock:  func(_ *MockAuthService) {},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "INVALID_REFRESH_TOKEN",
		},
		{
			name:        "service error returns error response without cookie",
			cookieValue: "refresh-token",
			setupMock: func(m *MockAuthService) {
				m.EXPECT().
					Refresh(mock.Anything, "refresh-token").
					Return(domain.TokenPair{}, errs.ErrInvalidRefreshToken).
					Once()
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "INVALID_REFRESH_TOKEN",
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

			handler.Refresh(rw, newRefreshRequest(t, tt.cookieValue))

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body errorResponseBody

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body.Code)
				is.Empty(rec.Header().Get("Set-Cookie"))

				return
			}

			var body refreshResponseBody

			must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
			is.Equal(tt.wantToken, body.AccessToken)

			cookies := rec.Result().Cookies()
			must.Len(cookies, 1)
			is.Equal("refresh_token", cookies[0].Name)
			is.Equal("new-refresh-token", cookies[0].Value)
			is.Equal("/", cookies[0].Path)
			is.True(cookies[0].HttpOnly)
			is.Equal(http.SameSiteLaxMode, cookies[0].SameSite)
			is.True(cookies[0].Expires.Equal(expiresAt))
			is.Equal(tt.cfg.CookieSecure, cookies[0].Secure)
			is.Equal(tt.cfg.CookieDomain, cookies[0].Domain)
			is.Len(rec.Header().Values("Set-Cookie"), 1)
		})
	}
}
