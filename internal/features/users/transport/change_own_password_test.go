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
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func newChangeOwnPasswordRequest(t *testing.T, userID uuid.UUID, body, contentType string) *http.Request {
	t.Helper()

	req := newTestJSONRequest(t, http.MethodPatch, "/users/me/password", body, contentType)

	return req.WithContext(reqctx.WithUserID(req.Context(), userID))
}

func TestChangeOwnPassword(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name        string
		body        string
		contentType string
		setupMock   func(m *MockUsersService)
		wantStatus  int
		wantCode    string
	}{
		{
			name:        "success changes own password",
			body:        `{"old_password":"oldSecret123","new_password":"newSecret123"}`,
			contentType: "application/json",
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					ChangeOwnPassword(mock.Anything, userID, "oldSecret123", "newSecret123").
					Return(nil).
					Once()
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:        "missing old password returns VALIDATION_FAILED",
			body:        `{"new_password":"newSecret123"}`,
			contentType: "application/json",
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name:        "missing new password returns VALIDATION_FAILED",
			body:        `{"old_password":"oldSecret123"}`,
			contentType: "application/json",
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name:        "new password too short returns VALIDATION_FAILED",
			body:        `{"old_password":"oldSecret123","new_password":"short"}`,
			contentType: "application/json",
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name:        "old password too short returns VALIDATION_FAILED",
			body:        `{"old_password":"short","new_password":"newSecret123"}`,
			contentType: "application/json",
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name:        "old password too long returns VALIDATION_FAILED",
			body:        `{"old_password":"` + strings.Repeat("a", 129) + `","new_password":"newSecret123"}`,
			contentType: "application/json",
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name:        "new password too long returns VALIDATION_FAILED",
			body:        `{"old_password":"oldSecret123","new_password":"` + strings.Repeat("a", 129) + `"}`,
			contentType: "application/json",
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name:        "wrong old password returns INVALID_CREDENTIALS",
			body:        `{"old_password":"wrong-pass","new_password":"newSecret123"}`,
			contentType: "application/json",
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					ChangeOwnPassword(mock.Anything, userID, "wrong-pass", "newSecret123").
					Return(errs.ErrInvalidCredentials).
					Once()
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_CREDENTIALS",
		},
		{
			name:        "malformed json returns INVALID_REQUEST_BODY",
			body:        `{"old_password":`,
			contentType: "application/json",
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST_BODY",
		},
		{
			name:        "unknown field returns INVALID_REQUEST_BODY",
			body:        `{"old_password":"oldSecret123","new_password":"newSecret123","extra":1}`,
			contentType: "application/json",
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST_BODY",
		},
		{
			name:        "body too large returns INVALID_REQUEST_BODY",
			body:        `{"old_password":"oldSecret123","new_password":"` + strings.Repeat("a", 1<<20) + `"}`,
			contentType: "application/json",
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST_BODY",
		},
		{
			name:        "invalid content type returns INVALID_CONTENT_TYPE",
			body:        `{"old_password":"oldSecret123","new_password":"newSecret123"}`,
			contentType: "text/plain",
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_CONTENT_TYPE",
		},
		{
			name:        "missing content type returns INVALID_CONTENT_TYPE",
			body:        `{"old_password":"oldSecret123","new_password":"newSecret123"}`,
			contentType: "",
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_CONTENT_TYPE",
		},
		{
			name:        "internal error is mapped to INTERNAL_ERROR",
			body:        `{"old_password":"oldSecret123","new_password":"newSecret123"}`,
			contentType: "application/json",
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					ChangeOwnPassword(mock.Anything, userID, "oldSecret123", "newSecret123").
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

			usersService := NewMockUsersService(t)
			tt.setupMock(usersService)

			handler := newUsersTestHandler(usersService)

			rec := httptest.NewRecorder()
			rw := core_http_response.NewResponseWriter(rec)

			handler.ChangeOwnPassword(rw, newChangeOwnPasswordRequest(t, userID, tt.body, tt.contentType))

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body errorResponseBody

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body.Code)
			}
		})
	}
}
