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
	"github.com/daniildddd/maestro/internal/core/logger"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func TestConnectorsHTTPHandler_RestartConnectorTask(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "restart task", func(a *gotest.Context) {
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
			{
				name: "service error returns REBALANCE_IN_PROGRESS",
				path: "/connectors/pg-connector/tasks/0/restart",
				setupMock: func(m *MockConnectorsService) {
					m.EXPECT().
						RestartTask(mock.Anything, "pg-connector", 0).
						Return(errs.ErrRebalanceInProgress).
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

				if tt.noTaskID {
					req = httptest.NewRequest(
						http.MethodPost,
						"/connectors/pg-connector/tasks/0/restart",
						http.NoBody,
					)
					req.SetPathValue("id", "pg-connector")
					req = req.WithContext(logger.ToContext(req.Context(), nopLogger()))
				}

				handler.RestartConnectorTask(rw, req)
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
		gotest.WithFeature("Restart task"),
		gotest.WithSeverity("critical"),
		gotest.WithOwner("qa-intern"),
	)
}
