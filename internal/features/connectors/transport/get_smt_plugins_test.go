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

func TestConnectorsHTTPHandler_GetSMTPlugins(t *testing.T) {
	t.Parallel()

	plugins := []domain.ConnectorPlugin{
		{
			ID: "io.debezium.transforms.ExtractNewRecordState",
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
			name: "success returns smt plugin ids",
			setupMock: func(m *MockConnectorsService) {
				m.EXPECT().
					GetSMTPlugins(mock.Anything).
					Return(plugins, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody: []transport.PluginListItemResponse{
				{
					ID: "io.debezium.transforms.ExtractNewRecordState",
				},
			},
		},
		{
			name: "service error returns KAFKA_CONNECT_UNAVAILABLE",
			setupMock: func(m *MockConnectorsService) {
				m.EXPECT().
					GetSMTPlugins(mock.Anything).
					Return(nil, errs.ErrKafkaConnectUnavailable).
					Once()
			},
			wantStatus: http.StatusServiceUnavailable,
			wantCode:   "KAFKA_CONNECT_UNAVAILABLE",
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

			handler.GetSMTPlugins(rw, newConnectorsRequest(t, http.MethodGet, "/smt-plugins", ""))

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
}
