package transport_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/logger"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func TestRestartConnectorTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		path       string
		setupMock  func(m *MockConnectorsService)
		wantStatus int
		wantCode   string
		noTaskID   bool
	}{
		{
			name: "success returns no content",
			path: "/connectors/pg-connector/tasks/0/restart",
			setupMock: func(m *MockConnectorsService) {
				m.EXPECT().
					RestartTask(mock.Anything, "pg-connector", 0).
					Return(nil).
					Once()
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "missing connector id returns error",
			path:       "/connectors/",
			setupMock:  func(_ *MockConnectorsService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_PATH_PARAM",
		},
		{
			name:       "missing task id returns error",
			path:       "/connectors/pg-connector/tasks/0/restart",
			setupMock:  func(_ *MockConnectorsService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_PATH_PARAM",
			noTaskID:   true,
		},
		{
			name: "service error returns NOT_FOUND",
			path: "/connectors/pg-connector/tasks/0/restart",
			setupMock: func(m *MockConnectorsService) {
				m.EXPECT().
					RestartTask(mock.Anything, "pg-connector", 0).
					Return(errs.ErrConnectorTaskNotFound).
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

			req := newConnectorsRequest(t, http.MethodPost, tt.path, "")

			if tt.noTaskID {
				req = httptest.NewRequest(http.MethodPost, "/connectors/pg-connector/tasks/0/restart", http.NoBody)
				req.SetPathValue("id", "pg-connector")
				req = req.WithContext(logger.ToContext(req.Context(), nopLogger()))
			}

			handler.RestartConnectorTask(rw, req)

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body errorResponseBody

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body.Code)

				return
			}

			is.Empty(rec.Body.Bytes())
		})
	}
}
