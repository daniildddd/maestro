package transport_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
	"github.com/daniildddd/maestro/internal/features/connectors/transport"
)

func TestConnectorsHTTPHandler_GetConnectorTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		path       string
		setupMock  func(m *MockConnectorsService)
		wantStatus int
		wantCode   string
		wantBody   transport.TaskDetailResponse
	}{
		{
			name: "success returns task detail with trace",
			path: "/connectors/pg-connector/tasks/0",
			setupMock: func(m *MockConnectorsService) {
				m.EXPECT().
					GetTaskByID(mock.Anything, "pg-connector", 0).
					Return(domain.Task{
						ID:       0,
						State:    "RUNNING",
						WorkerID: "worker-1",
						Trace:    new("worker died: oom"),
					}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody: transport.TaskDetailResponse{
				ID:       0,
				State:    "RUNNING",
				WorkerID: "worker-1",
				Trace:    new("worker died: oom"),
			},
		},
		{
			name:       "missing connector id returns error",
			path:       "/connectors/",
			setupMock:  func(_ *MockConnectorsService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_PATH_PARAM",
		},
		{
			name:       "malformed task id returns error",
			path:       "/connectors/pg-connector/tasks/abc",
			setupMock:  func(_ *MockConnectorsService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_PATH_PARAM",
		},
		{
			name: "service error returns NOT_FOUND",
			path: "/connectors/pg-connector/tasks/7",
			setupMock: func(m *MockConnectorsService) {
				m.EXPECT().
					GetTaskByID(mock.Anything, "pg-connector", 7).
					Return(domain.Task{}, errs.ErrConnectorTaskNotFound).
					Once()
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			must := require.New(t)

			connectorsService := NewMockConnectorsService(t)
			tt.setupMock(connectorsService)

			handler := newTestHandler(connectorsService)

			rec := httptest.NewRecorder()
			rw := core_http_response.NewResponseWriter(rec)

			req := newConnectorsRequest(t, http.MethodGet, tt.path, "")

			handler.GetConnectorTask(rw, req)

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body errorResponseBody

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body.Code)

				return
			}

			var body transport.TaskDetailResponse

			must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
			is.Equal(tt.wantBody, body)
		})
	}
}
