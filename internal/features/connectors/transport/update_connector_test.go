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

func TestUpdateConnector(t *testing.T) {
	t.Parallel()

	validBody := `{
		"config": {
			"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
			"database.hostname": "pg-2"
		}
	}`

	tests := []struct {
		name        string
		path        string
		body        string
		contentType *string
		setupMock   func(m *MockConnectorsService)
		wantStatus  int
		wantCode    string
		wantBody    transport.ConnectorDetailResponse
	}{
		{
			name: "success returns updated connector",
			path: "/connectors/pg-connector",
			body: validBody,
			setupMock: func(m *MockConnectorsService) {
				m.EXPECT().
					UpdateConnector(mock.Anything, "pg-connector", mock.Anything).
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
							Hostname: "pg-2",
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
					DatabaseHostname: "pg-2",
					DatabasePort:     "5432",
					DatabaseUser:     "debezium",
				},
			},
		},
		{
			name:       "missing connector id returns error",
			path:       "/connectors/",
			body:       validBody,
			setupMock:  func(_ *MockConnectorsService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_PATH_PARAM",
		},
		{
			name:       "empty config returns validation error",
			path:       "/connectors/pg-connector",
			body:       `{"config":{}}`,
			setupMock:  func(_ *MockConnectorsService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_FAILED",
		},
		{
			name:       "missing config returns validation error",
			path:       "/connectors/pg-connector",
			body:       `{}`,
			setupMock:  func(_ *MockConnectorsService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_FAILED",
		},
		{
			name:       "malformed json returns error",
			path:       "/connectors/pg-connector",
			body:       "{invalid",
			setupMock:  func(_ *MockConnectorsService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_REQUEST_BODY",
		},
		{
			name: "unknown field returns INVALID_REQUEST_BODY",
			path: "/connectors/pg-connector",
			body: `{
				"config": {"a": "b"},
				"extra": 1
			}`,
			setupMock:  func(_ *MockConnectorsService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_REQUEST_BODY",
		},
		{
			name: "missing content type returns INVALID_CONTENT_TYPE",
			path: "/connectors/pg-connector",
			body: `{
				"config": {"a": "b"}
			}`,
			contentType: strPtr(""),
			setupMock:   func(_ *MockConnectorsService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_CONTENT_TYPE",
		},
		{
			name: "wrong content type returns INVALID_CONTENT_TYPE",
			path: "/connectors/pg-connector",
			body: `{
				"config": {"a": "b"}
			}`,
			contentType: strPtr("text/plain"),
			setupMock:   func(_ *MockConnectorsService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_CONTENT_TYPE",
		},
		{
			name: "service error returns VALIDATION_FAILED",
			path: "/connectors/pg-connector",
			body: validBody,
			setupMock: func(m *MockConnectorsService) {
				m.EXPECT().
					UpdateConnector(mock.Anything, "pg-connector", mock.Anything).
					Return(domain.Connector{}, errs.ErrValidationFailed).
					Once()
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_FAILED",
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

			req := newConnectorsRequest(t, http.MethodPatch, tt.path, tt.body)
			if tt.contentType != nil {
				if *tt.contentType == "" {
					req.Header.Del("Content-Type")
				} else {
					req.Header.Set("Content-Type", *tt.contentType)
				}
			}

			handler.UpdateConnector(rw, req)

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
