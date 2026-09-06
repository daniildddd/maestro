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

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
	"github.com/daniildddd/maestro/internal/features/audit/transport"
)

//nolint:maintidx // table-driven test: complexity comes from per-case mock setups
func TestGetLogs(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	tests := []struct {
		name       string
		query      string
		setupMock  func(m *MockAuditService)
		wantStatus int
		wantCode   string
		wantBody   func(t *testing.T) transport.GetLogsResponse
	}{
		{
			name:  "success returns logs with meta",
			query: "page=1&limit=10",
			setupMock: func(m *MockAuditService) {
				m.EXPECT().
					GetLogs(
						mock.Anything,
						mock.MatchedBy(func(f *domain.AuditLogFilter) bool {
							return f.Page == 1 && f.Limit == 10
						}),
					).
					Return([]domain.AuditEvent{mustAuditEvent(t, id)}, false, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T) transport.GetLogsResponse {
				t.Helper()

				event := mustAuditEvent(t, id)

				return transport.GetLogsResponse{
					Data: []transport.AuditEventDTOResponse{
						{
							ID:        event.ID.String(),
							CreatedAt: event.CreatedAt,
							Action:    "auth.login",
							Outcome:   "success",
							Actor: &transport.ActorDTO{
								ID:       event.ActorID.String(),
								Username: event.ActorLogin,
							},
							Subject: transport.SubjectDTO{
								Type: domain.AuditSubjectUser,
								ID:   event.SubjectID,
								Name: event.SubjectName,
							},
							Request: &transport.RequestDTO{
								ID:        event.RequestID,
								IP:        event.IP,
								UserAgent: event.UserAgent,
							},
						},
					},
					Meta:    transport.PaginationMeta{Page: 1, Limit: 10},
					HasMore: false,
				}
			},
		},
		{
			name:  "has more flag from service is passed through",
			query: "page=1&limit=1",
			setupMock: func(m *MockAuditService) {
				m.EXPECT().
					GetLogs(mock.Anything, mock.Anything).
					Return([]domain.AuditEvent{mustAuditEvent(t, id)}, true, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T) transport.GetLogsResponse {
				t.Helper()

				event := mustAuditEvent(t, id)

				return transport.GetLogsResponse{
					Data: []transport.AuditEventDTOResponse{
						{
							ID:        event.ID.String(),
							CreatedAt: event.CreatedAt,
							Action:    "auth.login",
							Outcome:   "success",
							Actor: &transport.ActorDTO{
								ID:       event.ActorID.String(),
								Username: event.ActorLogin,
							},
							Subject: transport.SubjectDTO{
								Type: domain.AuditSubjectUser,
								ID:   event.SubjectID,
								Name: event.SubjectName,
							},
							Request: &transport.RequestDTO{
								ID:        event.RequestID,
								IP:        event.IP,
								UserAgent: event.UserAgent,
							},
						},
					},
					Meta:    transport.PaginationMeta{Page: 1, Limit: 1},
					HasMore: true,
				}
			},
		},
		{
			name:  "invalid page returns INVALID_QUERY_PARAM",
			query: "page=abc",
			setupMock: func(_ *MockAuditService) {
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_QUERY_PARAM",
		},
		{
			name:  "invalid limit returns INVALID_QUERY_PARAM",
			query: "limit=abc",
			setupMock: func(_ *MockAuditService) {
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_QUERY_PARAM",
		},
		{
			name:  "invalid action returns VALIDATION_FAILED",
			query: "page=1&limit=10&action=bogus.action",
			setupMock: func(_ *MockAuditService) {
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_FAILED",
		},
		{
			name:  "service error is wrapped",
			query: "page=1&limit=10",
			setupMock: func(m *MockAuditService) {
				m.EXPECT().
					GetLogs(mock.Anything, mock.Anything).
					Return(nil, false, errs.ErrInternal).
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

			auditService := NewMockAuditService(t)
			tt.setupMock(auditService)

			handler := newAuditTestHandler(auditService)

			rec := httptest.NewRecorder()
			rw := core_http_response.NewResponseWriter(rec)

			handler.GetLogs(rw, newAuditRequest(t, http.MethodGet, "/audit-logs?"+tt.query))

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body map[string]any

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body["code"])

				return
			}

			var body transport.GetLogsResponse

			must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
			is.Equal(tt.wantBody(t), body)
		})
	}
}

func TestGetLogsResponseDTO(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	is := assert.New(t)
	must := require.New(t)

	auditService := NewMockAuditService(t)
	auditService.EXPECT().
		GetLogs(mock.Anything, mock.Anything).
		Return([]domain.AuditEvent{mustAuditEvent(t, id)}, false, nil).
		Once()

	handler := newAuditTestHandler(auditService)

	rec := httptest.NewRecorder()
	rw := core_http_response.NewResponseWriter(rec)

	handler.GetLogs(rw, newAuditRequest(t, http.MethodGet, "/audit-logs?page=1&limit=10"))

	must.Equal(http.StatusOK, rec.Code)

	var body struct {
		Data []struct {
			ID      string `json:"id"`
			Action  string `json:"action"`
			Outcome string `json:"outcome"`
			Actor   *struct {
				ID       string `json:"id"`
				Username string `json:"username"`
			} `json:"actor"`
			Subject struct {
				Type string `json:"type"`
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"subject"`
			Request *struct {
				ID        string `json:"id"`
				IP        string `json:"ip"`
				UserAgent string `json:"user_agent"`
			} `json:"request"`
		} `json:"data"`
		Meta struct {
			Page  int `json:"page"`
			Limit int `json:"limit"`
		} `json:"meta"`
		HasMore bool `json:"has_more"`
	}

	must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))

	is.Len(body.Data, 1)
	is.Equal(id.String(), body.Data[0].ID)
	is.Equal("auth.login", body.Data[0].Action)
	is.Equal("success", body.Data[0].Outcome)

	is.NotNil(body.Data[0].Actor)
	is.Equal(id.String(), body.Data[0].Actor.ID)
	is.Equal("alice", body.Data[0].Actor.Username)

	is.Equal(domain.AuditSubjectUser, body.Data[0].Subject.Type)
	is.Equal(id.String(), body.Data[0].Subject.ID)
	is.Equal("alice", body.Data[0].Subject.Name)

	is.NotNil(body.Data[0].Request)
	is.Equal("req-"+id.String(), body.Data[0].Request.ID)
	is.Equal("203.0.113.7", body.Data[0].Request.IP)
	is.Equal("test-agent", body.Data[0].Request.UserAgent)

	is.Equal(1, body.Meta.Page)
	is.Equal(10, body.Meta.Limit)
	is.False(body.HasMore)
}
