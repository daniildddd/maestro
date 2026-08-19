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
	"github.com/daniildddd/maestro/internal/core/logger"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
	"github.com/daniildddd/maestro/internal/features/users/transport"
)

func newCreateUserRequest(
	t *testing.T,
	body string,
	contentType string,
) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return req.WithContext(logger.ToContext(req.Context(), nopLogger()))
}

func TestCreateUser(t *testing.T) {
	t.Parallel()

	createdAt := time.Now().Add(-time.Hour).Truncate(time.Second)
	userID := uuid.New()

	tests := []struct {
		name        string
		body        string
		contentType string
		setupMock   func(m *MockUsersService)
		wantStatus  int
		wantCode    string
		wantBody    transport.UserResponse
	}{
		{
			name:        "success creates user",
			body:        `{"username":"alice","password":"secret123","role":"admin"}`,
			contentType: "application/json",
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					CreateUser(
						mock.Anything,
						"alice",
						"secret123",
						"admin",
					).
					Return(
						mustNewUser(t,
							userID,
							"alice",
							"hash",
							"admin",
							createdAt,
							nil,
						),
						nil,
					).
					Once()
			},
			wantStatus: http.StatusCreated,
			wantBody: transport.UserResponse{
				ID:        userID.String(),
				Username:  "alice",
				Role:      "admin",
				CreatedAt: createdAt,
				UpdatedAt: nil,
			},
		},
		{
			name:        "invalid username rejected",
			body:        `{"username":"ab","password":"secret123","role":"admin"}`,
			contentType: "application/json",
			setupMock:   func(_ *MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name:        "password too short rejected",
			body:        `{"username":"alice","password":"1234567","role":"admin"}`,
			contentType: "application/json",
			setupMock:   func(_ *MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name:        "missing role rejected",
			body:        `{"username":"alice","password":"secret123"}`,
			contentType: "application/json",
			setupMock:   func(_ *MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name:        "malformed json rejected",
			body:        `{"username":`,
			contentType: "application/json",
			setupMock:   func(_ *MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST_BODY",
		},
		{
			name:        "unknown field rejected",
			body:        `{"username":"alice","password":"secret123","role":"admin","extra":1}`,
			contentType: "application/json",
			setupMock:   func(_ *MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST_BODY",
		},
		{
			name:        "missing content type rejected",
			body:        `{"username":"alice","password":"secret123","role":"admin"}`,
			contentType: "",
			setupMock:   func(_ *MockUsersService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_CONTENT_TYPE",
		},
		{
			name:        "username conflict returns 409",
			body:        `{"username":"alice","password":"secret123","role":"admin"}`,
			contentType: "application/json",
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					CreateUser(mock.Anything, "alice", "secret123", "admin").
					Return(domain.User{}, errs.ErrUsernameConflict).
					Once()
			},
			wantStatus: http.StatusConflict,
			wantCode:   "USERNAME_CONFLICT",
		},
		{
			name:        "service error returns 500",
			body:        `{"username":"alice","password":"secret123","role":"admin"}`,
			contentType: "application/json",
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					CreateUser(mock.Anything, "alice", "secret123", "admin").
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

			handler.CreateUser(
				rw,
				newCreateUserRequest(t, tt.body, tt.contentType),
			)

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
