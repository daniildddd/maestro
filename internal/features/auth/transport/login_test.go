package transport_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func newLoginRequest(t *testing.T, body, contentType string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/Login", strings.NewReader(body))

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return req.WithContext(logger.ToContext(req.Context(), nopLogger()))
}

func TestLogin(t *testing.T) {
	t.Parallel()

	expiresAt := time.Now().Add(time.Hour).Truncate(time.Second)

	tests := []struct {
		name        string
		contentType string
		body        string
		cfg         transport.Config
		setupMock   func(m *MockAuthService)
		wantStatus  int
		wantCode    string
		wantBody    transport.LoginUserResponse
	}{
		{
			name:        "valid credentials return tokens and set cookie",
			contentType: "application/json",
			body:        `{"username":"alice","password":"secret123"}`,
			cfg: transport.Config{
				CookieSecure: true,
				CookieDomain: "example.com",
			},
			setupMock: func(m *MockAuthService) {
				m.EXPECT().
					Login(mock.Anything, "alice", "secret123").
					Return(domain.TokenPair{
						AccessToken:  "access-token",
						RefreshToken: "refresh-token",
						ExpiresAt:    expiresAt,
					}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody: transport.LoginUserResponse{
				AccessToken: "access-token",
				Username:    "alice",
			},
		},
		{
			name:        "invalid JSON returns INVALID_REQUEST_BODY",
			contentType: "application/json",
			body:        `{invalid`,
			setupMock:   func(_ *MockAuthService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST_BODY",
		},
		{
			name:        "unknown field returns INVALID_REQUEST_BODY",
			contentType: "application/json",
			body:        `{"username":"alice","password":"secret123","extra":1}`,
			setupMock:   func(_ *MockAuthService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST_BODY",
		},
		{
			name:        "body too large returns INVALID_REQUEST_BODY",
			contentType: "application/json",
			body:        `{"username":"alice","password":"` + strings.Repeat("a", 1<<20) + `"}`,
			setupMock:   func(_ *MockAuthService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST_BODY",
		},
		{
			name:        "validation failure returns VALIDATION_FAILED",
			contentType: "application/json",
			body:        `{"username":"ab","password":"short"}`,
			setupMock:   func(_ *MockAuthService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name:        "username too long returns VALIDATION_FAILED",
			contentType: "application/json",
			body:        `{"username":"` + strings.Repeat("a", 33) + `","password":"secret123"}`,
			setupMock:   func(_ *MockAuthService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name:        "password too long returns VALIDATION_FAILED",
			contentType: "application/json",
			body:        `{"username":"alice","password":"` + strings.Repeat("a", 129) + `"}`,
			setupMock:   func(_ *MockAuthService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name:        "invalid content type returns INVALID_CONTENT_TYPE",
			contentType: "text/plain",
			body:        `{"username":"alice","password":"secret123"}`,
			setupMock:   func(_ *MockAuthService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_CONTENT_TYPE",
		},
		{
			name:        "missing content type returns INVALID_CONTENT_TYPE",
			contentType: "",
			body:        `{"username":"alice","password":"secret123"}`,
			setupMock:   func(_ *MockAuthService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_CONTENT_TYPE",
		},
		{
			name:        "service error returns INVALID_CREDENTIALS without cookie",
			contentType: "application/json",
			body:        `{"username":"alice","password":"wrong-password"}`,
			setupMock: func(m *MockAuthService) {
				m.EXPECT().
					Login(mock.Anything, "alice", "wrong-password").
					Return(domain.TokenPair{}, errs.ErrInvalidCredentials).
					Once()
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_CREDENTIALS",
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

			handler.Login(rw, newLoginRequest(t, tt.body, tt.contentType))

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body errorResponseBody

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body.Code)

				return
			}

			var body transport.LoginUserResponse

			must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
			is.Equal(tt.wantBody, body)

			cookies := rec.Result().Cookies()
			must.Len(cookies, 1)
			is.Equal("refresh_token", cookies[0].Name)
			is.Equal("refresh-token", cookies[0].Value)
			is.Equal("/", cookies[0].Path)
			is.True(cookies[0].HttpOnly)
			is.Equal(http.SameSiteLaxMode, cookies[0].SameSite)
			is.True(cookies[0].Expires.Equal(expiresAt))
			is.Equal(tt.cfg.CookieSecure, cookies[0].Secure)
			is.Equal(tt.cfg.CookieDomain, cookies[0].Domain)
		})
	}
}
