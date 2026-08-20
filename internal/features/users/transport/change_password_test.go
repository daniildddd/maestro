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
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func newChangePasswordRequest(t *testing.T, pathID, body, contentType string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(
		http.MethodPatch,
		"/users/{id}/password",
		strings.NewReader(body),
	)

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	req.SetPathValue("id", pathID)

	return req.WithContext(logger.ToContext(req.Context(), nopLogger()))
}

func TestChangePassword(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	longPass := strings.Repeat("a", 129)

	tests := []struct {
		name        string
		pathID      string
		contentType string
		body        string
		setupMock   func(m *MockUsersService)
		wantStatus  int
		wantCode    string
	}{
		{
			name:        "success changes password",
			pathID:      userID.String(),
			contentType: "application/json",
			body:        `{"new_password":"newSecret123"}`,
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					ChangePassword(mock.Anything, userID, "newSecret123").
					Return(nil).
					Once()
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:        "missing password returns VALIDATION_FAILED",
			pathID:      userID.String(),
			contentType: "application/json",
			body:        `{}`,
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name:        "password too short returns VALIDATION_FAILED",
			pathID:      userID.String(),
			contentType: "application/json",
			body:        `{"new_password":"short"}`,
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name:        "password too long returns VALIDATION_FAILED",
			pathID:      userID.String(),
			contentType: "application/json",
			body:        `{"new_password":"` + longPass + `"}`,
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name:        "malformed json returns INVALID_REQUEST_BODY",
			pathID:      userID.String(),
			contentType: "application/json",
			body:        `{"new_password":`,
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST_BODY",
		},
		{
			name:        "unknown field returns INVALID_REQUEST_BODY",
			pathID:      userID.String(),
			contentType: "application/json",
			body:        `{"new_password":"newSecret123","extra":1}`,
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST_BODY",
		},



		{
			name:        "invalid path id returns INVALID_PATH_PARAM",
			pathID:      "not-a-uuid",
			contentType: "application/json",
			body:        `{"new_password":"newSecret123"}`,
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_PATH_PARAM",
		},
		{
			name:        "user not found returns USER_NOT_FOUND",
			pathID:      userID.String(),
			contentType: "application/json",
			body:        `{"new_password":"newSecret123"}`,
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					ChangePassword(mock.Anything, userID, "newSecret123").
					Return(errs.ErrUserNotFound).
					Once()
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "USER_NOT_FOUND",
		},
		{
			name:        "internal error is mapped to INTERNAL_ERROR",
			pathID:      userID.String(),
			contentType: "application/json",
			body:        `{"new_password":"newSecret123"}`,
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					ChangePassword(mock.Anything, userID, "newSecret123").
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

			handler.ChangePassword(rw, newChangePasswordRequest(t, tt.pathID, tt.body, tt.contentType))

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body errorResponseBody

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body.Code)
			}
		})
	}
}
