package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/features/connectors/service"
)

func TestConnectorsServiceGetConnectorPluginSchema(t *testing.T) {
	t.Parallel()

	allFields := []domain.ConnectorPluginField{
		{
			Name:       "database.hostname",
			Importance: domain.PluginImportanceHigh,
			Required:   true,
		},
		{
			Name:       "database.port",
			Importance: domain.PluginImportanceHigh,
		},
		{
			Name:       "plugin.name",
			Importance: domain.PluginImportanceMedium,
		},
		{
			Name:       "internal.clock",
			Importance: domain.PluginImportanceLow,
		},
	}

	tests := []struct {
		name      string
		pluginID  string
		filter    *domain.ConnectorPluginSchemaFilter
		setupMock func(kc *MockKafkaConnect)
		want      domain.ConnectorPluginSchema
		wantIs    error
	}{
		{
			name:     "high filter keeps high fields",
			pluginID: "io.debezium.connector.postgresql.PostgresConnector",
			filter:   mustSchemaFilter(t, "high"),
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorPluginSchemaWithValues(mock.Anything, "io.debezium.connector.postgresql.PostgresConnector").
					Return(domain.ConnectorPluginSchema{Fields: allFields}, nil).
					Once()
			},
			want: domain.ConnectorPluginSchema{Fields: allFields[:2]},
		},
		{
			name:     "medium filter keeps high and medium fields",
			pluginID: "io.debezium.connector.postgresql.PostgresConnector",
			filter:   mustSchemaFilter(t, "medium"),
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorPluginSchemaWithValues(mock.Anything, "io.debezium.connector.postgresql.PostgresConnector").
					Return(domain.ConnectorPluginSchema{Fields: allFields}, nil).
					Once()
			},
			want: domain.ConnectorPluginSchema{Fields: allFields[:3]},
		},
		{
			name:     "all filter keeps everything",
			pluginID: "io.debezium.connector.postgresql.PostgresConnector",
			filter:   mustSchemaFilter(t, "all"),
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorPluginSchemaWithValues(mock.Anything, "io.debezium.connector.postgresql.PostgresConnector").
					Return(domain.ConnectorPluginSchema{Fields: allFields}, nil).
					Once()
			},
			want: domain.ConnectorPluginSchema{Fields: allFields},
		},
		{
			name:     "plugin not found maps to ErrConnectorPluginNotFound",
			pluginID: "unknown.Connector",
			filter:   mustSchemaFilter(t, "high"),
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorPluginSchemaWithValues(mock.Anything, "unknown.Connector").
					Return(domain.ConnectorPluginSchema{}, domain.ErrConnectorPluginNotFound).
					Once()
			},
			wantIs: errs.ErrConnectorPluginNotFound,
		},
		{
			name:     "connect unavailable maps to ErrKafkaConnectUnavailable",
			pluginID: "io.debezium.connector.postgresql.PostgresConnector",
			filter:   mustSchemaFilter(t, "high"),
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorPluginSchemaWithValues(mock.Anything, "io.debezium.connector.postgresql.PostgresConnector").
					Return(domain.ConnectorPluginSchema{}, domain.ErrKafkaConnectUnavailable).
					Once()
			},
			wantIs: errs.ErrKafkaConnectUnavailable,
		},
		{
			name:     "unexpected error passes through wrapped",
			pluginID: "io.debezium.connector.postgresql.PostgresConnector",
			filter:   mustSchemaFilter(t, "high"),
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorPluginSchemaWithValues(mock.Anything, "io.debezium.connector.postgresql.PostgresConnector").
					Return(domain.ConnectorPluginSchema{}, errSourceDown).
					Once()
			},
			wantIs: errSourceDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			kc := NewMockKafkaConnect(t)
			tt.setupMock(kc)

			svc := service.NewConnectorsService(kc, noopAuditor{})

			got, err := svc.GetConnectorPluginSchema(context.Background(), tt.pluginID, tt.filter)

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			must.Equal(tt.want, got)
		})
	}
}

func mustSchemaFilter(t *testing.T, importance string) *domain.ConnectorPluginSchemaFilter {
	t.Helper()

	filter, err := domain.NewConnectorPluginSchemaFilter(importance)
	if err != nil {
		t.Fatalf("build schema filter: %v", err)
	}

	return filter
}
