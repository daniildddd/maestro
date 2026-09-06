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
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func TestDeleteConnector(t *testing.T) {
	t.Parallel()

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
			name: "service error returns mapped error",
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			must := require.New(t)

			connectorsService := NewMockConnectorsService(t)
			tt.setupMock(connectorsService)

			handler := newTestHandler(connectorsService)

			rec := httptest.NewRecorder()
			rw := core_http_response.NewResponseWriter(rec)

			req := newConnectorsRequest(t, http.MethodDelete, tt.path, "")

			handler.DeleteConnector(rw, req)

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
