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

func TestConnectorsHTTPHandler_GetConnectorPluginSchema(t *testing.T) {
	t.Parallel()

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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			must := require.New(t)

			connectorsService := NewMockConnectorsService(t)
			tt.setupMock(connectorsService)

			handler := newTestHandler(connectorsService)

			rec := httptest.NewRecorder()
			rw := core_http_response.NewResponseWriter(rec)

			req := newConnectorsRequest(t, http.MethodGet, tt.path, "")

			handler.GetConnectorPluginSchema(rw, req)

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
}
