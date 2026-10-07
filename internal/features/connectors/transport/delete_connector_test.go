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

	"github.com/daniildddd/maestro/internal/core/errs"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func TestConnectorsHTTPHandler_DeleteConnector(t *testing.T) { //nolint:paralleltest // allure gotest.Test owns subtest execution
	gotest.Test(t, "delete connector", func(a *gotest.Context) {
		tests := []struct {
			name       string
			path       string
			setupMock  func(m *MockConnectorsService)
			wantStatus int
			wantCode   string
		}{
			{
				name: "success returns no content",
				path: "/connectors/pg-connector",
				setupMock: func(m *MockConnectorsService) {
					m.EXPECT().
						Delete(mock.Anything, "pg-connector").
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
				name: "service error returns CONNECTOR_NOT_FOUND",
				path: "/connectors/pg-connector",
				setupMock: func(m *MockConnectorsService) {
					m.EXPECT().
						Delete(mock.Anything, "pg-connector").
						Return(errs.ErrConnectorNotFound).
						Once()
				},
				wantStatus: http.StatusNotFound,
				wantCode:   "CONNECTOR_NOT_FOUND",
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

				req := newConnectorsRequest(a.T(), http.MethodDelete, tt.path, "")

				handler.DeleteConnector(rw, req)
				a.Attachment("response.json", rec.Body.Bytes(), "application/json")

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
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Delete connector"),
		gotest.WithSeverity("critical"),
		gotest.WithOwner("qa-intern"),
	)
}
