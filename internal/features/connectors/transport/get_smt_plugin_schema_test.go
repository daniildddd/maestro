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

func TestGetSMTPluginSchema(t *testing.T) {
	t.Parallel()

	fields := []domain.ConnectorPluginField{
		{
			Name:       "transforms",
			Type:       "STRING",
			Importance: domain.PluginImportanceHigh,
			Required:   true,
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
			name: "success returns smt schema fields",
			path: "/smt-plugins/{id}/config?importance=high",
			setupMock: func(m *MockConnectorsService) {
				m.EXPECT().
					GetSMTPluginSchema(mock.Anything, "{id}", mock.Anything).
					Return(domain.ConnectorPluginSchema{Fields: fields}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody: transport.PluginSchemaResponse{
				Fields: []transport.PluginSchemaFieldResponse{
					{
						Name:       "transforms",
						Type:       "STRING",
						Importance: "HIGH",
						Required:   true,
					},
				},
			},
		},
		{
			name:       "missing connector id returns error",
			path:       "/smt-plugins//config?importance=high",
			setupMock:  func(_ *MockConnectorsService) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_PATH_PARAM",
		},
		{
			name: "invalid importance returns error",
			path: "/smt-plugins/{id}/config?importance=bogus",
			setupMock: func(_ *MockConnectorsService) {
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_QUERY_PARAM",
		},
		{
			name: "service error returns NOT_FOUND",
			path: "/smt-plugins/{id}/config?importance=high",
			setupMock: func(m *MockConnectorsService) {
				m.EXPECT().
					GetSMTPluginSchema(mock.Anything, "{id}", mock.Anything).
					Return(domain.ConnectorPluginSchema{}, errs.ErrSmtPluginNotFound).
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

			handler.GetSMTPluginSchema(rw, req)

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
