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

func newDeleteMeRequest(t *testing.T, userID uuid.UUID, body, contentType string) *http.Request {
	t.Helper()

	req := newTestJSONRequest(t, http.MethodDelete, "/users/me", body, contentType)

	return req.WithContext(reqctx.WithUserID(req.Context(), userID))
}

func TestDeleteMe(t *testing.T) {
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
			name:        "success deletes own account",
			body:        `{"password":"secret123"}`,
			contentType: "application/json",
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					DeleteMe(mock.Anything, userID, "secret123").
					Return(nil).
					Once()
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:        "wrong password returns INVALID_CREDENTIALS",
			body:        `{"password":"wrong-pass"}`,
			contentType: "application/json",
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
			name:        "missing password returns VALIDATION_FAILED",
			body:        `{}`,
			contentType: "application/json",
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},


		{
			name:        "malformed json returns INVALID_REQUEST_BODY",
			body:        `{"password":`,
			contentType: "application/json",
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST_BODY",
		},
		{
			name:        "unknown field returns INVALID_REQUEST_BODY",
			body:        `{"password":"secret123","extra":1}`,
			contentType: "application/json",
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST_BODY",
		},
		{
			name:        "body too large returns INVALID_REQUEST_BODY",
			body:        `{"password":"` + strings.Repeat("a", 1<<20) + `"}`,
			contentType: "application/json",
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST_BODY",
		},
		{
			name:        "invalid content type returns INVALID_CONTENT_TYPE",
			body:        `{"password":"secret123"}`,
			contentType: "text/plain",
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_CONTENT_TYPE",
		},
		{
			name:        "missing content type returns INVALID_CONTENT_TYPE",
			body:        `{"password":"secret123"}`,
			contentType: "",
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_CONTENT_TYPE",
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

			handler.DeleteMe(rw, newDeleteMeRequest(t, userID, tt.body, tt.contentType))

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body errorResponseBody

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body.Code)
			}
		})
	}
}
