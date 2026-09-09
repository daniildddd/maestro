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

func TestGetConnectors(t *testing.T) {
	t.Parallel()

	pgConnector := domain.Connector{
		Name:       "pg-connector",
		PluginType: "io.debezium.connector.postgresql.PostgresConnector",
		Status:     domain.ConnectorStatusRunning,
		TasksCount: 1,
	}

	tests := []struct {
		name       string
		query      string
		setupMock  func(m *MockConnectorsService)
		wantStatus int
		wantCode   string
		wantBody   transport.ConnectorListResponse
	}{
		{
			name:  "success returns connectors with meta and total",
			query: "page=1&limit=10",
			setupMock: func(m *MockConnectorsService) {
				m.EXPECT().
					GetConnectors(mock.Anything, mock.Anything).
					Return([]domain.Connector{pgConnector}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody: transport.ConnectorListResponse{
				Data: []transport.ConnectorResponse{
					{
						Name:       "pg-connector",
						PluginType: "io.debezium.connector.postgresql.PostgresConnector",
						Status:     "running",
						TasksCount: 1,
					},
				},
				Meta:  transport.Meta{Page: 1, Limit: 10},
				Total: 1,
			},
		},
		{
			name:  "invalid page returns error",
			query: "page=abc",
			setupMock: func(_ *MockConnectorsService) {
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_QUERY_PARAM",
		},
		{
			name:  "invalid limit returns error",
			query: "limit=abc",
			setupMock: func(_ *MockConnectorsService) {
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_QUERY_PARAM",
		},
		{
			name:  "invalid status filter returns error",
			query: "page=1&limit=10&status=bogus",
			setupMock: func(_ *MockConnectorsService) {
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_QUERY_PARAM",
		},
		{
			name:  "service error returns KAFKA_CONNECT_UNAVAILABLE",
			query: "page=1&limit=10",
			setupMock: func(m *MockConnectorsService) {
				m.EXPECT().
					GetConnectors(mock.Anything, mock.Anything).
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

			handler.GetConnectors(rw, newConnectorsRequest(t, http.MethodGet, "/connectors?"+tt.query, ""))

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body errorResponseBody

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body.Code)

				return
			}

			var body transport.ConnectorListResponse

			must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
			is.Equal(tt.wantBody, body)
		})
	}
}
