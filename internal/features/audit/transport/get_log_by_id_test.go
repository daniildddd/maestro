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
)

func TestGetLogByID(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	tests := []struct {
		name       string
		pathID     string
		setupMock  func(m *MockAuditService)
		wantStatus int
		wantCode   string
	}{
		{
			name:   "success returns event by id",
			pathID: id.String(),
			setupMock: func(m *MockAuditService) {
				m.EXPECT().
					GetLogByID(mock.Anything, id).
					Return(mustAuditEvent(t, id), nil).
					Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "malformed uuid returns INVALID_PATH_PARAM",
			pathID:     "not-a-uuid",
			setupMock:  func(_ *MockAuditService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_PATH_PARAM",
		},
		{
			name:   "service error is wrapped",
			pathID: id.String(),
			setupMock: func(m *MockAuditService) {
				m.EXPECT().
					GetLogByID(mock.Anything, id).
					Return(domain.AuditEvent{}, errs.ErrAuditLogNotFound).
					Once()
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "AUDIT_LOG_NOT_FOUND",
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

			handler.GetLogByID(rw, newAuditRequest(t, http.MethodGet, "/audit-logs/"+tt.pathID))

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body errorResponseBody

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body.Code)

				return
			}

			var body struct {
				ID     string `json:"id"`
				Action string `json:"action"`
			}

			must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
			is.Equal(id.String(), body.ID)
			is.Equal("auth.login", body.Action)
		})
	}
}
