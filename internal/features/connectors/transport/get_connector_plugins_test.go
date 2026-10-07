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

func TestConnectorsHTTPHandler_GetConnectorPlugins(t *testing.T) { //nolint:paralleltest // allure gotest.Test owns subtest execution
	gotest.Test(t, "get connector plugins", func(a *gotest.Context) {
		plugins := []domain.ConnectorPlugin{
			{
				ID: "io.debezium.connector.postgresql.PostgresConnector",
			},
			{
				ID: "io.debezium.connector.mysql.MySqlConnector",
			},
		}

		tests := []struct {
			name       string
			setupMock  func(m *MockConnectorsService)
			wantStatus int
			wantCode   string
			wantBody   []transport.PluginListItemResponse
		}{
			{
				name: "success returns plugin ids",
				setupMock: func(m *MockConnectorsService) {
					m.EXPECT().
						GetConnectorPlugins(mock.Anything).
						Return(plugins, nil).
						Once()
				},
				wantStatus: http.StatusOK,
				wantBody: []transport.PluginListItemResponse{
					{
						ID: "io.debezium.connector.postgresql.PostgresConnector",
					},
					{
						ID: "io.debezium.connector.mysql.MySqlConnector",
					},
				},
			},
			{
				name: "service error returns KAFKA_CONNECT_UNAVAILABLE",
				setupMock: func(m *MockConnectorsService) {
					m.EXPECT().
						GetConnectorPlugins(mock.Anything).
						Return(nil, errs.ErrKafkaConnectUnavailable).
						Once()
				},
				wantStatus: http.StatusServiceUnavailable,
				wantCode:   "KAFKA_CONNECT_UNAVAILABLE",
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

				handler.GetConnectorPlugins(
					rw,
					newConnectorsRequest(a.T(), http.MethodGet, "/connector-plugins", ""),
				)
				a.Attachment("response.json", rec.Body.Bytes(), "application/json")

				must.Equal(tt.wantStatus, rec.Code)

				if tt.wantCode != "" {
					var body errorResponseBody

					must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
					is.Equal(tt.wantCode, body.Code)

					return
				}

				var body []transport.PluginListItemResponse

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantBody, body)
			})
		}
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Get connector plugins"),
		gotest.WithSeverity("normal"),
		gotest.WithOwner("qa-intern"),
	)
}
