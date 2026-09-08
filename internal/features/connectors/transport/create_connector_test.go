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

func TestCreateConnector(t *testing.T) {
	t.Parallel()

	validBody := `{
		"name": "pg-connector",
		"config": {
			"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
			"database.hostname": "pg-1"
		}
	}`

	tests := []struct {
		name       string
		body       string
		setupMock  func(m *MockConnectorsService)
		wantStatus int
		wantCode   string
		wantBody   transport.ConnectorCreateResponse
	}{
		{
			name: "success returns created connector",
			body: validBody,
			setupMock: func(m *MockConnectorsService) {
				m.EXPECT().
					CreateConnector(mock.Anything, "pg-connector", mock.Anything).
					Return(domain.Connector{
						Name:       "pg-connector",
						PluginType: "io.debezium.connector.postgresql.PostgresConnector",
						Config: domain.SourceConfig{
							Hostname:   "pg-1",
							Port:       "5432",
							User:       "debezium",
							DBName:     strPtr("postgres"),
							PluginName: strPtr("pgoutput"),
						},
						Status:     domain.ConnectorStatusStarting,
						TasksCount: 0,
						Tasks:      []domain.Task{},
					}, nil).
					Once()
			},
			wantStatus: http.StatusCreated,
			wantBody: transport.ConnectorCreateResponse{
				Name:       "pg-connector",
				PluginType: "io.debezium.connector.postgresql.PostgresConnector",
				Status:     "starting",
				Config: transport.SourceConfigDTO{
					DatabaseHostname: "pg-1",
					DatabasePort:     "5432",
					DatabaseUser:     "debezium",
					DatabaseDBName:   strPtr("postgres"),
					PluginName:       strPtr("pgoutput"),
				},
				TasksCount: 0,
				Tasks:      []transport.TaskResponse{},
			},
		},
		{
			name: "empty name returns validation error",
			body: `{
				"name": "",
				"config": {"a": "b"}
			}`,
			setupMock:  func(_ *MockConnectorsService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_FAILED",
		},
		{
			name: "empty config returns validation error",
			body: `{
				"name": "pg-connector",
				"config": {}
			}`,
			setupMock:  func(_ *MockConnectorsService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_FAILED",
		},
		{
			name:       "malformed json returns error",
			body:       "{invalid",
			setupMock:  func(_ *MockConnectorsService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_REQUEST_BODY",
		},
		{
			name: "service error returns mapped error",
			body: validBody,
			setupMock: func(m *MockConnectorsService) {
				m.EXPECT().
					CreateConnector(mock.Anything, "pg-connector", mock.Anything).
					Return(domain.Connector{}, errs.ErrConnectorAlreadyExists).
					Once()
			},
			wantStatus: http.StatusConflict,
			wantCode:   "CONNECTOR_ALREADY_EXISTS",
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

			handler.CreateConnector(rw, newConnectorsRequest(t, http.MethodPost, "/connectors", tt.body))

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body errorResponseBody

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body.Code)

				return
			}

			var body transport.ConnectorCreateResponse

			must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
			is.Equal(tt.wantBody, body)
		})
	}
}
