package transport_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/logger"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func newDeleteUserRequest(t *testing.T, pathID string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodDelete, "/users/{id}", http.NoBody)
	req.SetPathValue("id", pathID)

	return req.WithContext(logger.ToContext(req.Context(), nopLogger()))
}

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name       string
		pathID     string
		setupMock  func(m *MockUsersService)
		wantStatus int
		wantCode   string
	}{
		{
			name:   "success deletes user",
			pathID: userID.String(),
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					DeleteUser(mock.Anything, userID).
					Return(nil).
					Once()
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "malformed uuid returns INVALID_PATH_PARAM",
			pathID: "not-a-uuid",
			setupMock: func(_ *MockUsersService) {
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_PATH_PARAM",
		},
		{
			name:   "empty path id returns INVALID_PATH_PARAM",
			pathID: "",
			setupMock: func(_ *MockUsersService) {
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_PATH_PARAM",
		},
		{
			name:       "nil uuid path id returns INVALID_PATH_PARAM",
			pathID:     "00000000-0000-0000-0000-000000000000",
			setupMock:  func(*MockUsersService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_PATH_PARAM",
		},
		{
			name:   "user not found returns USER_NOT_FOUND",
			pathID: userID.String(),
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					DeleteUser(mock.Anything, userID).
					Return(errs.ErrUserNotFound).
					Once()
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "USER_NOT_FOUND",
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

			handler.DeleteUser(rw, newDeleteUserRequest(t, tt.pathID))

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body errorResponseBody

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body.Code)
			}
		})
	}
}
