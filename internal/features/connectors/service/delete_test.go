package service_test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/features/connectors/service"
)

func TestConnectorsServiceDelete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupMock func(kc *MockKafkaConnect)
		wantIs    error
	}{
		{
			name: "delete succeeds",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorByID(mock.Anything, "pg-connector").
					Return(domain.Connector{Name: "pg-connector", PluginType: "postgres", Status: "running"}, nil).
					Once()

				kc.EXPECT().
					Delete(mock.Anything, "pg-connector").
					Return(nil).
					Once()
			},
		},
		{
			name: "missing connector maps to app error",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorByID(mock.Anything, "pg-connector").
					Return(domain.Connector{}, domain.ErrConnectorNotFound).
					Once()
			},
			wantIs: errs.ErrConnectorNotFound,
		},
		{
			name: "connect unavailable maps to app error",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorByID(mock.Anything, "pg-connector").
					Return(domain.Connector{}, domain.ErrKafkaConnectUnavailable).
					Once()
			},
			wantIs: errs.ErrKafkaConnectUnavailable,
		},
		{
			name: "rebalance maps to app error",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorByID(mock.Anything, "pg-connector").
					Return(domain.Connector{}, domain.ErrRebalanceInProgress).
					Once()
			},
			wantIs: errs.ErrRebalanceInProgress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			kc := NewMockKafkaConnect(t)
			tt.setupMock(kc)

			svc := service.NewConnectorsService(kc, noopAuditor{})

			err := svc.Delete(testCtx(), "pg-connector")

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
		})
	}
}
