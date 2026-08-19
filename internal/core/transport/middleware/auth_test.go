package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap/zapcore"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	core_errs "github.com/daniildddd/maestro/internal/core/errs"
	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/security/access"
	"github.com/daniildddd/maestro/internal/core/transport/middleware"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func TestAuth(t *testing.T) {
	t.Parallel()

	validUserID := uuid.New()
	validRole := "admin"

	tests := []struct {
		name       string
		authHeader string
		verifier   *fakeTokenVerifier
		wantStatus int
		wantCode   string
		wantNext   bool
		wantUserID uuid.UUID
		wantRole   string
	}{
		{
			name:       "valid token passes through with user context",
			authHeader: "Bearer valid",
			verifier:   &fakeTokenVerifier{user: access.AuthUser{UserID: validUserID, Role: validRole}},
			wantStatus: http.StatusOK,
			wantNext:   true,
			wantUserID: validUserID,
			wantRole:   validRole,
		},
		{
			name:       "missing Authorization header returns 401",
			authHeader: "",
			verifier:   &fakeTokenVerifier{},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "ACCESS_TOKEN_MISSING",
			wantNext:   false,
		},
		{
			name:       "Authorization without Bearer prefix returns 401",
			authHeader: "Basic xyz",
			verifier:   &fakeTokenVerifier{},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "ACCESS_TOKEN_MISSING",
			wantNext:   false,
		},
		{
			name:       "verify error returns 401",
			authHeader: "Bearer invalid",
			verifier:   &fakeTokenVerifier{err: core_errs.ErrAccessTokenInvalid},
			wantStatus: http.StatusUnauthorized,
			wantNext:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)

			log, recordedLogs := newObservableLogger(t, zapcore.InfoLevel)

			var (
				capturedUserID uuid.UUID
				capturedRole   string
				nextCalled     bool
			)

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				capturedUserID = reqctx.UserID(r.Context())
				capturedRole = reqctx.Role(r.Context())
				core_logger.FromContext(r.Context()).Info("test") //nolint:contextcheck // test handler: request context from httptest.NewRequest is not derived from a parent
				w.WriteHeader(http.StatusOK)
			})

			headers := http.Header{}

			if tt.authHeader != "" {
				headers.Set("Authorization", tt.authHeader)
			}

			req := newTestRequest(t, http.MethodGet, "/", headers)
			req = req.WithContext(core_logger.ToContext(req.Context(), log))
			rec := httptest.NewRecorder()
			rw := core_http_response.NewResponseWriter(rec)

			middleware.Auth(tt.verifier)(nextHandler).ServeHTTP(rw, req)

			must.Equal(tt.wantStatus, rec.Code)
			must.Equal(tt.wantNext, nextCalled)

			if tt.wantCode != "" {
				must.Contains(rec.Body.String(), tt.wantCode)
			}

			if !tt.wantNext {
				must.False(rw.AuthDone)

				return
			}

			must.Equal(tt.wantUserID, capturedUserID)
			must.Equal(tt.wantRole, capturedRole)
			must.True(rw.AuthDone)
			must.Equal(tt.wantUserID, rw.UserID)
			must.Equal(tt.wantRole, rw.Role)

			logs := recordedLogs.All()
			must.Len(logs, 1)

			ctxMap := logs[0].ContextMap()
			must.Equal(tt.wantUserID.String(), ctxMap["user_id"])
			must.Equal(tt.wantRole, ctxMap["role"])
		})
	}

	t.Run("panics without logger in context", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		tv := &fakeTokenVerifier{user: access.AuthUser{UserID: uuid.New(), Role: "admin"}}
		nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			t.Fatal("next handler must not be called")
		})

		req := newTestRequest(t, http.MethodGet, "/", http.Header{
			"Authorization": []string{"Bearer valid"},
		})
		rec := httptest.NewRecorder()

		must.Panics(func() {
			middleware.Auth(tv)(nextHandler).ServeHTTP(rec, req)
		})
	})
}
