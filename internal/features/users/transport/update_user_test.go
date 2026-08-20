package transport_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
	"github.com/daniildddd/maestro/internal/features/users/transport"
)

func newUpdateUserRequest(t *testing.T, pathID, body, contentType string) *http.Request {
	t.Helper()

	req := newTestJSONRequest(t, http.MethodPatch, "/users/{id}", body, contentType)
	req.SetPathValue("id", pathID)

	return req
}

func TestUpdateUser(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	createdAt := time.Now().Add(-time.Hour).Truncate(time.Second)
	updatedAt := time.Now().Add(-time.Minute).Truncate(time.Second)

	tests := []struct {
		name        string
		pathID      string
		contentType string
		body        string
		setupMock   func(m *MockUsersService)
		wantStatus  int
		wantCode    string
		wantBody    transport.UserResponse
	}{
		{
			name:        "success updates username",
			pathID:      userID.String(),
			contentType: "application/json",
			body:        `{"username":"updateduser"}`,
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					UpdateUser(mock.Anything, userID, "updateduser").
					Return(mustNewUser(t,
						userID,
						"updateduser",
						"hash",
						"user",
						createdAt,
						&updatedAt,
					), nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody: transport.UserResponse{
				ID:        userID.String(),
				Username:  "updateduser",
				Role:      "user",
				CreatedAt: createdAt,
				UpdatedAt: &updatedAt,
			},
		},
		{
			name:        "missing username returns VALIDATION_FAILED",
			pathID:      userID.String(),
			contentType: "application/json",
			body:        `{}`,
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name:        "username too short returns VALIDATION_FAILED",
			pathID:      userID.String(),
			contentType: "application/json",
			body:        `{"username":"ab"}`,
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name:        "username too long returns VALIDATION_FAILED",
			pathID:      userID.String(),
			contentType: "application/json",
			body:        `{"username":"` + strings.Repeat("a", 33) + `"}`,
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name:        "malformed json returns INVALID_REQUEST_BODY",
			pathID:      userID.String(),
			contentType: "application/json",
			body:        `{"username":`,
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST_BODY",
		},
		{
			name:        "unknown field returns INVALID_REQUEST_BODY",
			pathID:      userID.String(),
			contentType: "application/json",
			body:        `{"username":"updateduser","extra":1}`,
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST_BODY",
		},
		{
			name:        "body too large returns INVALID_REQUEST_BODY",
			pathID:      userID.String(),
			contentType: "application/json",
			body:        `{"username":"` + strings.Repeat("a", 1<<20) + `"}`,
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST_BODY",
		},
		{
			name:        "invalid content type returns INVALID_CONTENT_TYPE",
			pathID:      userID.String(),
			contentType: "text/plain",
			body:        `{"username":"updateduser"}`,
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_CONTENT_TYPE",
		},
		{
			name:        "missing content type returns INVALID_CONTENT_TYPE",
			pathID:      userID.String(),
			contentType: "",
			body:        `{"username":"updateduser"}`,
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_CONTENT_TYPE",
		},
		{
			name:        "invalid path id returns INVALID_PATH_PARAM",
			pathID:      "not-a-uuid",
			contentType: "application/json",
			body:        `{"username":"updateduser"}`,
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_PATH_PARAM",
		},
		{
			name:        "nil uuid path id returns INVALID_PATH_PARAM",
			pathID:      "00000000-0000-0000-0000-000000000000",
			contentType: "application/json",
			body:        `{"username":"updateduser"}`,
			setupMock:   func(*MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_PATH_PARAM",
		},
		{
			name:        "user not found returns USER_NOT_FOUND",
			pathID:      userID.String(),
			contentType: "application/json",
			body:        `{"username":"updateduser"}`,
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					UpdateUser(mock.Anything, userID, "updateduser").
					Return(domain.User{}, errs.ErrUserNotFound).
					Once()
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "USER_NOT_FOUND",
		},
		{
			name:        "internal error is mapped to INTERNAL_ERROR",
			pathID:      userID.String(),
			contentType: "application/json",
			body:        `{"username":"updateduser"}`,
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					UpdateUser(mock.Anything, userID, "updateduser").
					Return(domain.User{}, errs.ErrInternal).
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

			handler.UpdateUser(rw, newUpdateUserRequest(t, tt.pathID, tt.body, tt.contentType))

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body errorResponseBody

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body.Code)

				return
			}

			var body transport.UserResponse

			must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
			is.Equal(tt.wantBody, body)
		})
	}
}
