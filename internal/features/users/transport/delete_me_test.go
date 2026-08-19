package transport_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func newDeleteMeRequest(t *testing.T, userID uuid.UUID, body string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(
		http.MethodDelete,
		"/users/me",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	ctx := logger.ToContext(req.Context(), nopLogger())
	ctx = reqctx.WithUserID(ctx, userID)

	return req.WithContext(ctx)
}

func TestDeleteMe(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name       string
		body       string
		setupMock  func(m *MockUsersService)
		wantStatus int
		wantCode   string
	}{
		{
			name: "success deletes own account",
			body: `{"password":"secret123"}`,
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					DeleteMe(mock.Anything, userID, "secret123").
					Return(nil).
					Once()
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "wrong password returns INVALID_CREDENTIALS",
			body: `{"password":"wrong-pass"}`,
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					DeleteMe(mock.Anything, userID, "wrong-pass").
					Return(errs.ErrInvalidCredentials).
					Once()
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_CREDENTIALS",
		},
		{
			name: "missing password returns VALIDATION_FAILED",
			body: `{}`,
			setupMock: func(_ *MockUsersService) {
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			usersService := NewMockUsersService(t)
			tt.setupMock(usersService)

			handler := newUsersTestHandler(usersService)

			rec := httptest.NewRecorder()
			rw := core_http_response.NewResponseWriter(rec)

			handler.DeleteMe(rw, newDeleteMeRequest(t, userID, tt.body))

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body errorResponseBody

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body.Code)
			}
		})
	}
}
