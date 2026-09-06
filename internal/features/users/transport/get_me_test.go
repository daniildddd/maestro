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
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
	"github.com/daniildddd/maestro/internal/features/users/transport"
)

func newGetMeRequest(t *testing.T, userID uuid.UUID) *http.Request {
	t.Helper()

	req := newTestJSONRequest(t, http.MethodGet, "/users/me", "", "")
	req = req.WithContext(reqctx.WithUserID(req.Context(), userID))

	return req
}

func TestGetMe(t *testing.T) {
	t.Parallel()

	createdAt := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	updatedAt := time.Now().UTC().Add(-time.Minute).Truncate(time.Second)
	userID := uuid.New()

	tests := []struct {
		name       string
		userID     uuid.UUID
		setupMock  func(m *MockUsersService)
		wantStatus int
		wantCode   string
		wantBody   transport.GetMeResponse
	}{
		{
			name:   "success returns authenticated user",
			userID: userID,
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
			wantBody: transport.GetMeResponse{
				ID:        userID.String(),
				Username:  "alice",
				Role:      "admin",
				CreatedAt: createdAt,
				UpdatedAt: &updatedAt,
			},
		},
		{
			name:   "user with nil updated_at returns null",
			userID: userID,
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
			wantBody: transport.GetMeResponse{
				ID:        userID.String(),
				Username:  "bob",
				Role:      "user",
				CreatedAt: createdAt,
				UpdatedAt: nil,
			},
		},
		{
			name:   "service error is wrapped",
			userID: userID,
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					GetUserByID(mock.Anything, userID).
					Return(domain.User{}, errs.ErrUserNotFound).
					Once()
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "USER_NOT_FOUND",
		},
		{
			name:   "service error returns INTERNAL_ERROR",
			userID: userID,
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					GetUserByID(mock.Anything, userID).
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

			handler.GetMe(rw, newGetMeRequest(t, tt.userID))

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body errorResponseBody

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body.Code)

				return
			}

			var body transport.GetMeResponse

			must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
			is.Equal(tt.wantBody, body)
		})
	}
}
