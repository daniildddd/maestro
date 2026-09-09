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

func TestConnectorsServiceGetConnectorPlugins(t *testing.T) {
	t.Parallel()

	plugins := []domain.ConnectorPlugin{
		{
			ID: "io.debezium.connector.postgresql.PostgresConnector",
		},
		{
			ID: "io.debezium.connector.mysql.MySqlConnector",
		},
	}

	tests := []struct {
		name      string
		setupMock func(kc *MockKafkaConnect)
		want      []domain.ConnectorPlugin
		wantIs    error
	}{
		{
			name: "returns plugins as-is",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorPlugins(mock.Anything).
					Return(plugins, nil).
					Once()
			},
			want: plugins,
		},
		{
			name: "kafka connect unavailable maps to ErrKafkaConnectUnavailable",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorPlugins(mock.Anything).
					Return(nil, domain.ErrKafkaConnectUnavailable).
					Once()
			},
			wantIs: errs.ErrKafkaConnectUnavailable,
		},
		{
			name: "unexpected error passes through",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorPlugins(mock.Anything).
					Return(nil, errSourceDown).
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

			got, err := svc.GetConnectorPlugins(context.Background())

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			must.Equal(tt.want, got)
		})
	}
}
