package service_test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/features/connectors/service"
)

func TestCreateConnector(t *testing.T) {
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
			name: "success returns created connector",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					CreateConnector(mock.Anything, "pg-connector", config).
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
			name: "existing connector maps to conflict error",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					CreateConnector(mock.Anything, "pg-connector", config).
					Return(domain.Connector{}, domain.ErrConnectorAlreadyExists).
					Once()
			},
			wantIs: errs.ErrConnectorAlreadyExists,
		},
		{
			name: "rebalance maps to app error",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					CreateConnector(mock.Anything, "pg-connector", config).
					Return(domain.Connector{}, domain.ErrRebalanceInProgress).
					Once()
			},
			wantIs: errs.ErrRebalanceInProgress,
		},
		{
			name: "invalid config maps to validation error",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					CreateConnector(mock.Anything, "pg-connector", config).
					Return(domain.Connector{}, domain.ErrInvalidConnectorConfig).
					Once()
			},
			wantIs: errs.ErrValidationFailed,
		},
		{
			name: "connect unavailable maps to app error",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					CreateConnector(mock.Anything, "pg-connector", config).
					Return(domain.Connector{}, domain.ErrKafkaConnectUnavailable).
					Once()
			},
			wantIs: errs.ErrKafkaConnectUnavailable,
		},
		{
			name: "unexpected error passes through wrapped",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					CreateConnector(mock.Anything, "pg-connector", config).
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

			connector, err := svc.CreateConnector(testCtx(), "pg-connector", config)

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			must.Equal(tt.want, connector)
		})
	}
}
