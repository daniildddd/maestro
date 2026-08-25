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
	"github.com/daniildddd/maestro/internal/core/logger"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
	"github.com/daniildddd/maestro/internal/features/users/transport"
)

func newGetUsersRequest(t *testing.T, query string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/users?"+query, http.NoBody)

	return req.WithContext(logger.ToContext(req.Context(), nopLogger()))
}

func TestGetUsers(t *testing.T) {
	t.Parallel()

	createdAt := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	updatedAt := time.Now().UTC().Add(-time.Minute).Truncate(time.Second)
	userID := uuid.New()

	tests := []struct {
		name       string
		query      string
		setupMock  func(m *MockUsersService)
		wantStatus int
		wantCode   string
		wantBody   transport.GetUsersResponse
	}{
		{
			name:  "success returns users with meta",
			query: "page=2&limit=10",
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					GetUsers(
						mock.Anything,
						mock.MatchedBy(func(f *domain.UserFilter) bool {
							return f.Page == 2 && f.Limit == 10
						}),
					).
					Return([]domain.User{
						mustNewUser(t,
							userID,
							"alice",
							"hash",
							"admin",
							createdAt,
							&updatedAt,
						),
					}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody: transport.GetUsersResponse{
				Data: []transport.UserDTOResponse{
					{
						ID:        userID.String(),
						Username:  "alice",
						Role:      "admin",
						CreatedAt: createdAt,
						UpdatedAt: &updatedAt,
					},
				},
				Meta: transport.PaginationMeta{
					Page:  2,
					Limit: 10,
				},
			},
		},
		{
			name:  "user with nil updated_at returns null",
			query: "page=2&limit=10",
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					GetUsers(mock.Anything, mock.Anything).
					Return([]domain.User{
						mustNewUser(t,
							userID,
							"bob",
							"hash",
							"user",
							createdAt,
							nil,
						),
					}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody: transport.GetUsersResponse{
				Data: []transport.UserDTOResponse{
					{
						ID:        userID.String(),
						Username:  "bob",
						Role:      "user",
						CreatedAt: createdAt,
						UpdatedAt: nil,
					},
				},
				Meta: transport.PaginationMeta{
					Page:  2,
					Limit: 10,
				},
			},
		},
		{
			name:  "defaults applied when page and limit absent",
			query: "",
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					GetUsers(
						mock.Anything,
						mock.MatchedBy(func(f *domain.UserFilter) bool {
							return f.Page == 1 && f.Limit == 20
						}),
					).
					Return([]domain.User{
						mustNewUser(t,
							userID,
							"carol",
							"hash",
							"user",
							createdAt,
							nil,
						),
					}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody: transport.GetUsersResponse{
				Data: []transport.UserDTOResponse{
					{
						ID:        userID.String(),
						Username:  "carol",
						Role:      "user",
						CreatedAt: createdAt,
						UpdatedAt: nil,
					},
				},
				Meta: transport.PaginationMeta{
					Page:  1,
					Limit: 20,
				},
			},
		},
		{
			name:  "empty result returns empty data",
			query: "page=2&limit=10",
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					GetUsers(mock.Anything, mock.Anything).
					Return(nil, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody: transport.GetUsersResponse{
				Data: []transport.UserDTOResponse{},
				Meta: transport.PaginationMeta{
					Page:  2,
					Limit: 10,
				},
			},
		},
		{
			name:       "invalid page returns INVALID_QUERY_PARAM",
			query:      "page=abc",
			setupMock:  func(_ *MockUsersService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_QUERY_PARAM",
		},
		{
			name:       "invalid limit returns INVALID_QUERY_PARAM",
			query:      "limit=abc",
			setupMock:  func(_ *MockUsersService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_QUERY_PARAM",
		},
		{
			name:       "invalid role returns VALIDATION_FAILED",
			query:      "role=superadmin",
			setupMock:  func(_ *MockUsersService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_FAILED",
		},
		{
			name:       "invalid username returns VALIDATION_FAILED",
			query:      "username=bad%20name%21",
			setupMock:  func(_ *MockUsersService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_FAILED",
		},
		{
			name:  "service error is wrapped",
			query: "",
			setupMock: func(m *MockUsersService) {
				m.EXPECT().
					GetUsers(mock.Anything, mock.Anything).
					Return(nil, errs.ErrInternal).
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

			handler.GetUsers(rw, newGetUsersRequest(t, tt.query))

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body errorResponseBody

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body.Code)

				return
			}

			var body transport.GetUsersResponse

			must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
			is.Equal(tt.wantBody, body)
		})
	}
}
