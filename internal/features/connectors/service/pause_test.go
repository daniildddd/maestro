package service_test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/features/connectors/service"
)

func TestConnectorsService_PauseConnector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupMock func(kc *MockKafkaConnect)
		want      domain.Connector
		wantIs    error
	}{
		{
			name: "success returns paused connector",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					PauseConnector(mock.Anything, "pg-connector").
					Return(domain.Connector{
						Name:       "pg-connector",
						PluginType: "io.debezium.connector.postgresql.PostgresConnector",
						Status:     domain.ConnectorStatusRunning,
					}, nil).
					Once()
			},
			want: domain.Connector{
				Name:       "pg-connector",
				PluginType: "io.debezium.connector.postgresql.PostgresConnector",
				Status:     domain.ConnectorStatusRunning,
			},
		},
		{
			name: "missing connector maps to ErrConnectorNotFound",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					PauseConnector(mock.Anything, "pg-connector").
					Return(domain.Connector{}, domain.ErrConnectorNotFound).
					Once()
			},
			wantIs: errs.ErrConnectorNotFound,
		},
		{
			name: "rebalance maps to ErrRebalanceInProgress",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					PauseConnector(mock.Anything, "pg-connector").
					Return(domain.Connector{}, domain.ErrRebalanceInProgress).
					Once()
			},
			wantIs: errs.ErrRebalanceInProgress,
		},
		{
			name: "connect unavailable maps to ErrKafkaConnectUnavailable",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					PauseConnector(mock.Anything, "pg-connector").
					Return(domain.Connector{}, domain.ErrKafkaConnectUnavailable).
					Once()
			},
			wantIs: errs.ErrKafkaConnectUnavailable,
		},
		{
			name: "unexpected error passes through wrapped",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					PauseConnector(mock.Anything, "pg-connector").
					Return(domain.Connector{}, errSourceDown).
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

			connector, err := svc.PauseConnector(testCtx(), "pg-connector")

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			must.Equal(tt.want, connector)
		})
	}

	t.Run("records connector state in audit", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		dbName := "shop"
		pluginName := "pgoutput"

		kc := NewMockKafkaConnect(t)
		kc.EXPECT().
			PauseConnector(mock.Anything, "pg-connector").
			Return(domain.Connector{
				Name:       "pg-connector",
				PluginType: "io.debezium.connector.postgresql.PostgresConnector",
				Status:     domain.ConnectorStatusRunning,
				Config: domain.SourceConfig{
					Hostname:   "pg-1",
					Port:       "5432",
					User:       "debezium",
					DBName:     &dbName,
					PluginName: &pluginName,
				},
			}, nil).
			Once()

		auditor := &capturingAuditor{}
		svc := service.NewConnectorsService(kc, auditor)

		_, err := svc.PauseConnector(testCtx(), "pg-connector")
		must.NoError(err)
		must.Len(auditor.recorded, 1)

		state := auditor.recorded[0].StateAfter

		must.Equal(map[string]any{
			"name":              "pg-connector",
			"plugin_type":       "io.debezium.connector.postgresql.PostgresConnector",
			"status":            domain.ConnectorStatusRunning,
			"database.hostname": "pg-1",
			"database.port":     "5432",
			"database.user":     "debezium",
			"database.dbname":   "shop",
			"plugin.name":       "pgoutput",
		}, state)
	})
}
