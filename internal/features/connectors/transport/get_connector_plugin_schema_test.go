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

func TestGetConnectorPluginSchema(t *testing.T) {
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
		query      string
		setupMock  func(m *MockConnectorsService)
		wantStatus int
		wantCode   string
		wantBody   transport.PluginSchemaResponse
		rawRequest bool
	}{
		{
			name:  "success returns schema fields",
			query: "importance=all",
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
			name:       "missing path id returns error",
			setupMock:  func(_ *MockConnectorsService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_PATH_PARAM",
			rawRequest: true,
		},
		{
			name:  "invalid importance returns error",
			query: "importance=bogus",
			setupMock: func(_ *MockConnectorsService) {
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_QUERY_PARAM",
		},
		{
			name:  "service error returns mapped error",
			query: "importance=all",
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

			req := newConnectorsRequest(t, http.MethodGet, "/connector-plugins/{id}/config?"+tt.query, "")
			if tt.rawRequest {
				req = newConnectorsRequestWithoutPathValues(t, http.MethodGet, "/connector-plugins/{id}/config?"+tt.query, "")
			}

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
