package transport_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gotest "github.com/allure-framework/allure-go/commons/gotest" //nolint:revive // package clause is allure, gotest keeps call sites readable
	allureassert "github.com/allure-framework/allure-go/testify/assert"
	allurerequire "github.com/allure-framework/allure-go/testify/require"
	"github.com/stretchr/testify/mock"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
	"github.com/daniildddd/maestro/internal/features/connectors/transport"
)

func TestConnectorsHTTPHandler_RestartConnector(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "restart connector", func(a *gotest.Context) {
		tests := []struct {
			name       string
			path       string
			setupMock  func(m *MockConnectorsService)
			wantStatus int
			wantCode   string
			wantBody   transport.ConnectorDetailResponse
		}{
			{
				name: "success without query flags restarts connector",
				path: "/connectors/pg-connector/restart",
				setupMock: func(m *MockConnectorsService) {
					m.EXPECT().
						RestartConnector(mock.Anything, "pg-connector", false, false).
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
								Trace:    new("trace-1"),
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
				name: "query flags are passed to service",
				path: "/connectors/pg-connector/restart?include_tasks=true&only_failed=true",
				setupMock: func(m *MockConnectorsService) {
					m.EXPECT().
						RestartConnector(mock.Anything, "pg-connector", true, true).
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
								Trace:    new("trace-1"),
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
				name:       "missing connector id returns error",
				path:       "/connectors/",
				setupMock:  func(_ *MockConnectorsService) {},
				wantStatus: http.StatusBadRequest,
				wantCode:   "INVALID_PATH_PARAM",
			},
			{
				name:       "invalid include_tasks returns error",
				path:       "/connectors/pg-connector/restart?include_tasks=yes",
				setupMock:  func(_ *MockConnectorsService) {},
				wantStatus: http.StatusBadRequest,
				wantCode:   "INVALID_QUERY_PARAM",
			},
			{
				name:       "invalid only_failed returns error",
				path:       "/connectors/pg-connector/restart?only_failed=yes",
				setupMock:  func(_ *MockConnectorsService) {},
				wantStatus: http.StatusBadRequest,
				wantCode:   "INVALID_QUERY_PARAM",
			},
			{
				name: "service error returns CONNECTOR_NOT_FOUND",
				path: "/connectors/pg-connector/restart?include_tasks=true",
				setupMock: func(m *MockConnectorsService) {
					m.EXPECT().
						RestartConnector(mock.Anything, "pg-connector", true, false).
						Return(domain.Connector{}, errs.ErrConnectorNotFound).
						Once()
				},
				wantStatus: http.StatusNotFound,
				wantCode:   "CONNECTOR_NOT_FOUND",
			},
			{
				name: "service error returns REBALANCE_IN_PROGRESS",
				path: "/connectors/pg-connector/restart?include_tasks=true",
				setupMock: func(m *MockConnectorsService) {
					m.EXPECT().
						RestartConnector(mock.Anything, "pg-connector", true, false).
						Return(domain.Connector{}, errs.ErrRebalanceInProgress).
						Once()
				},
				wantStatus: http.StatusServiceUnavailable,
				wantCode:   "REBALANCE_IN_PROGRESS",
			},
		}

		for _, tt := range tests {
			a.Step(tt.name, func(a *gotest.Context) {
				is := allureassert.New(a)
				must := allurerequire.New(a)

				connectorsService := NewMockConnectorsService(a.T())
				tt.setupMock(connectorsService)

				handler := newTestHandler(connectorsService)

				rec := httptest.NewRecorder()
				rw := core_http_response.NewResponseWriter(rec)

				req := newConnectorsRequest(a.T(), http.MethodPost, tt.path, "")

				handler.RestartConnector(rw, req)
				a.Attachment("response.json", rec.Body.Bytes(), "application/json")

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
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Restart connector"),
		gotest.WithSeverity("critical"),
		gotest.WithOwner("qa-intern"),
	)
}
