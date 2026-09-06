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

func TestPauseConnector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		path       string
		setupMock  func(m *MockConnectorsService)
		wantStatus int
		wantCode   string
		wantBody   transport.ConnectorDetailResponse
	}{
		{
			name: "success returns paused connector",
			path: "/connectors/pg-connector/pause",
			setupMock: func(m *MockConnectorsService) {
				m.EXPECT().
					PauseConnector(mock.Anything, "pg-connector").
					Return(domain.Connector{
						Name:       "pg-connector",
						PluginType: "io.debezium.connector.postgresql.PostgresConnector",
						Status:     domain.ConnectorStatusRunning,
						WorkerID:   "worker-1",
						TasksCount: 1,
						Tasks: []domain.Task{{
							ID:       0,
							State:    "RUNNING",
							WorkerID: "worker-1",
							Trace:    strPtr("trace-1"),
						}},
						Config: domain.SourceConfig{
							Hostname: "pg-1",
							Port:     "5432",
							User:     "debezium",
						},
					}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody: transport.ConnectorDetailResponse{
				Name:       "pg-connector",
				PluginType: "io.debezium.connector.postgresql.PostgresConnector",
				Status:     "running",
				WorkerID:   "worker-1",
				TasksCount: 1,
				Tasks: []transport.TaskResponse{{
					ID:       0,
					State:    "RUNNING",
					WorkerID: "worker-1",
				}},
				Config: transport.SourceConfigDTO{
					DatabaseHostname: "pg-1",
					DatabasePort:     "5432",
					DatabaseUser:     "debezium",
				},
			},
		},
		{
			name:       "missing path id returns error",
			path:       "/connectors/",
			setupMock:  func(_ *MockConnectorsService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_PATH_PARAM",
		},
		{
			name: "service error returns mapped error",
			path: "/connectors/pg-connector/pause",
			setupMock: func(m *MockConnectorsService) {
				m.EXPECT().
					PauseConnector(mock.Anything, "pg-connector").
					Return(domain.Connector{}, errs.ErrRebalanceInProgress).
					Once()
			},
			wantStatus: http.StatusServiceUnavailable,
			wantCode:   "REBALANCE_IN_PROGRESS",
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

			handler.PauseConnector(rw, req)

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body errorResponseBody

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body.Code)

				return
			}

			var body transport.ConnectorDetailResponse

			must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
			is.Equal(tt.wantBody, body)
		})
	}
}
