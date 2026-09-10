package service_test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/features/connectors/service"
)

func TestConnectorsService_UpdateConnector(t *testing.T) {
	t.Parallel()

	config := map[string]string{
		"connector.class":   postgresClass,
		"database.hostname": "pg-1",
	}

	tests := []struct {
		name      string
		setupMock func(kc *MockKafkaConnect)
		want      domain.Connector
		wantIs    error
	}{
		{
			name: "success returns updated connector",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorByID(mock.Anything, "pg-connector").
					Return(domain.Connector{
						Name:       "pg-connector",
						PluginType: "io.debezium.connector.postgresql.PostgresConnector",
						Status:     domain.ConnectorStatusRunning,
					}, nil).
					Once()

				kc.EXPECT().
					UpdateConnector(mock.Anything, "pg-connector", config).
					Return(domain.Connector{
						Name:       "pg-connector",
						PluginType: "io.debezium.connector.postgresql.PostgresConnector",
						Status:     domain.ConnectorStatusRunning,
					}, nil).
					Once()

				kc.EXPECT().
					GetConnectorPluginSchema(mock.Anything, postgresClass).
					Return(domain.ConnectorPluginSchema{Fields: []domain.ConnectorPluginField{
						{
							Name: "database.password",
							Type: "PASSWORD",
						},
					}}, nil).
					Once()
			},
			want: domain.Connector{
				Name:       "pg-connector",
				PluginType: "io.debezium.connector.postgresql.PostgresConnector",
				Status:     domain.ConnectorStatusRunning,
			},
		},
		{
			name: "missing connector before update maps to ErrConnectorNotFound",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorByID(mock.Anything, "pg-connector").
					Return(domain.Connector{}, domain.ErrConnectorNotFound).
					Once()
			},
			wantIs: errs.ErrConnectorNotFound,
		},
		{
			name: "missing connector on update maps to ErrConnectorNotFound",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorByID(mock.Anything, "pg-connector").
					Return(domain.Connector{
						Name:       "pg-connector",
						PluginType: "io.debezium.connector.postgresql.PostgresConnector",
						Status:     domain.ConnectorStatusRunning,
					}, nil).
					Once()

				kc.EXPECT().
					UpdateConnector(mock.Anything, "pg-connector", config).
					Return(domain.Connector{}, domain.ErrConnectorNotFound).
					Once()
			},
			wantIs: errs.ErrConnectorNotFound,
		},
		{
			name: "rebalance maps to ErrRebalanceInProgress",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorByID(mock.Anything, "pg-connector").
					Return(domain.Connector{
						Name:       "pg-connector",
						PluginType: "io.debezium.connector.postgresql.PostgresConnector",
						Status:     domain.ConnectorStatusRunning,
					}, nil).
					Once()

				kc.EXPECT().
					UpdateConnector(mock.Anything, "pg-connector", config).
					Return(domain.Connector{}, domain.ErrRebalanceInProgress).
					Once()
			},
			wantIs: errs.ErrRebalanceInProgress,
		},
		{
			name: "invalid config maps to validation error",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorByID(mock.Anything, "pg-connector").
					Return(domain.Connector{
						Name:       "pg-connector",
						PluginType: "io.debezium.connector.postgresql.PostgresConnector",
						Status:     domain.ConnectorStatusRunning,
					}, nil).
					Once()

				kc.EXPECT().
					UpdateConnector(mock.Anything, "pg-connector", config).
					Return(domain.Connector{}, domain.ErrInvalidConnectorConfig).
					Once()
			},
			wantIs: errs.ErrValidationFailed,
		},
		{
			name: "connect unavailable maps to ErrKafkaConnectUnavailable",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorByID(mock.Anything, "pg-connector").
					Return(domain.Connector{
						Name:       "pg-connector",
						PluginType: "io.debezium.connector.postgresql.PostgresConnector",
						Status:     domain.ConnectorStatusRunning,
					}, nil).
					Once()

				kc.EXPECT().
					UpdateConnector(mock.Anything, "pg-connector", config).
					Return(domain.Connector{}, domain.ErrKafkaConnectUnavailable).
					Once()
			},
			wantIs: errs.ErrKafkaConnectUnavailable,
		},
		{
			name: "unexpected error passes through wrapped",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorByID(mock.Anything, "pg-connector").
					Return(domain.Connector{
						Name:       "pg-connector",
						PluginType: "io.debezium.connector.postgresql.PostgresConnector",
						Status:     domain.ConnectorStatusRunning,
					}, nil).
					Once()

				kc.EXPECT().
					UpdateConnector(mock.Anything, "pg-connector", config).
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

			connector, err := svc.UpdateConnector(testCtx(), "pg-connector", config)

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

		config := map[string]string{
			"connector.class":   postgresClass,
			"database.hostname": "pg-1",
			"database.password": "s3cret",
		}

		kc := NewMockKafkaConnect(t)
		kc.EXPECT().
			GetConnectorByID(mock.Anything, "pg-connector").
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

		kc.EXPECT().
			UpdateConnector(mock.Anything, "pg-connector", config).
			Return(domain.Connector{
				Name:       "pg-connector",
				PluginType: "io.debezium.connector.postgresql.PostgresConnector",
				Status:     domain.ConnectorStatusRunning,
			}, nil).
			Once()

		kc.EXPECT().
			GetConnectorPluginSchema(mock.Anything, postgresClass).
			Return(domain.ConnectorPluginSchema{
				Fields: []domain.ConnectorPluginField{
					{
						Name: "database.password",
						Type: "PASSWORD",
					},
				},
			}, nil).
			Once()

		auditor := &capturingAuditor{}
		svc := service.NewConnectorsService(kc, auditor)

		_, err := svc.UpdateConnector(testCtx(), "pg-connector", config)
		must.NoError(err)
		must.Len(auditor.recorded, 1)

		must.Equal(map[string]any{
			"name":              "pg-connector",
			"plugin_type":       "io.debezium.connector.postgresql.PostgresConnector",
			"status":            domain.ConnectorStatusRunning,
			"database.hostname": "pg-1",
			"database.port":     "5432",
			"database.user":     "debezium",
			"database.dbname":   "shop",
			"plugin.name":       "pgoutput",
		}, auditor.recorded[0].StateBefore)

		must.Equal(map[string]any{
			"connector.class":   postgresClass,
			"database.hostname": "pg-1",
			"database.password": "***",
		}, auditor.recorded[0].StateAfter)
	})
}
