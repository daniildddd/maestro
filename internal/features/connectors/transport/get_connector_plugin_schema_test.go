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

func TestConnectorsHTTPHandler_GetConnectorPluginSchema(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "get connector plugin schema", func(a *gotest.Context) {
		fields := []domain.ConnectorPluginField{
			{
				Name:       "database.hostname",
				Type:       "STRING",
				Importance: domain.PluginImportanceHigh,
				Required:   true,
			},
			{
				Name:       "database.password",
				Type:       "PASSWORD",
				Importance: domain.PluginImportanceHigh,
			},
			{
				Name:       "internal.option",
				Type:       "INT",
				Importance: domain.PluginImportanceLow,
			},
		}

		tests := []struct {
			name       string
			path       string
			setupMock  func(m *MockConnectorsService)
			wantStatus int
			wantCode   string
			wantBody   transport.PluginSchemaResponse
		}{
			{
				name: "success returns schema fields",
				path: "/connector-plugins/{id}/config?importance=all",
				setupMock: func(m *MockConnectorsService) {
					m.EXPECT().
						GetConnectorPluginSchema(mock.Anything, "{id}", mock.Anything).
						Return(domain.ConnectorPluginSchema{Fields: fields}, nil).
						Once()
				},
				wantStatus: http.StatusOK,
				wantBody: transport.PluginSchemaResponse{
					Fields: []transport.PluginSchemaFieldResponse{
						{
							Name:       "database.hostname",
							Type:       "STRING",
							Importance: "HIGH",
							Required:   true,
						},
						{
							Name:       "database.password",
							Type:       "PASSWORD",
							Importance: "HIGH",
						},
						{
							Name:       "internal.option",
							Type:       "INT",
							Importance: "LOW",
						},
					},
				},
			},
			{
				name:       "missing connector id returns error",
				path:       "/connector-plugins//config?importance=all",
				setupMock:  func(_ *MockConnectorsService) {},
				wantStatus: http.StatusBadRequest,
				wantCode:   "INVALID_PATH_PARAM",
			},
			{
				name: "invalid importance returns error",
				path: "/connector-plugins/{id}/config?importance=bogus",
				setupMock: func(_ *MockConnectorsService) {
				},
				wantStatus: http.StatusBadRequest,
				wantCode:   "INVALID_QUERY_PARAM",
			},
			{
				name: "service error returns NOT_FOUND",
				path: "/connector-plugins/{id}/config?importance=all",
				setupMock: func(m *MockConnectorsService) {
					m.EXPECT().
						GetConnectorPluginSchema(mock.Anything, "{id}", mock.Anything).
						Return(domain.ConnectorPluginSchema{}, errs.ErrConnectorPluginNotFound).
						Once()
				},
				wantStatus: http.StatusNotFound,
				wantCode:   "NOT_FOUND",
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

				req := newConnectorsRequest(a.T(), http.MethodGet, tt.path, "")

				handler.GetConnectorPluginSchema(rw, req)
				a.Attachment("response.json", rec.Body.Bytes(), "application/json")

				must.Equal(tt.wantStatus, rec.Code)

				if tt.wantCode != "" {
					var body errorResponseBody

					must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
					is.Equal(tt.wantCode, body.Code)

					return
				}

				var body transport.PluginSchemaResponse

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantBody, body)
			})
		}
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Get connector plugin schema"),
		gotest.WithSeverity("normal"),
		gotest.WithOwner("qa-intern"),
	)
}
