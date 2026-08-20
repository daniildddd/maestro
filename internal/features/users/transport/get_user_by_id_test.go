package transport_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func newGetUserByIDRequest(t *testing.T, pathID string) *http.Request {
	t.Helper()

	req := newTestJSONRequest(t, http.MethodGet, "/users/{id}", "", "")
	req.SetPathValue("id", pathID)

	return req
}

func TestGetUserByID(t *testing.T) {
	t.Parallel()

	createdAt := time.Now().Add(-time.Hour).Truncate(time.Second)
	updatedAt := time.Now().Add(-time.Minute).Truncate(time.Second)
	userID := uuid.New()

	tests := []struct {
		name       string
		pathID     string
		setupMock  func(m *MockUsersService)
		wantStatus int
		wantCode   string
		wantBody   transport.UserResponse
	}{
		{
			name:   "success returns user",
			pathID: userID.String(),
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					GetUserByID(mock.Anything, userID).
					Return(mustNewUser(t,
						userID,
						"alice",
						"hash",
						"admin",
						createdAt,
						&updatedAt,
					), nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody: transport.UserResponse{
				ID:        userID.String(),
				Username:  "alice",
				Role:      "admin",
				CreatedAt: createdAt,
				UpdatedAt: &updatedAt,
			},
		},
		{
			name:   "user with nil updated_at returns null",
			pathID: userID.String(),
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					GetUserByID(mock.Anything, userID).
					Return(mustNewUser(t,
						userID,
						"bob",
						"hash",
						"user",
						createdAt,
						nil,
					), nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody: transport.UserResponse{
				ID:        userID.String(),
				Username:  "bob",
				Role:      "user",
				CreatedAt: createdAt,
				UpdatedAt: nil,
			},
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
					GetUserByID(mock.Anything, userID).
					Return(domain.User{}, errs.ErrUserNotFound).
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

			handler.GetUserByID(rw, newGetUserByIDRequest(t, tt.pathID))

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
