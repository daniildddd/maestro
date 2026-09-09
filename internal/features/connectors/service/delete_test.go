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
			name: "success deletes existing connector",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorByID(mock.Anything, "pg-connector").
					Return(domain.Connector{
						Name:       "pg-connector",
						PluginType: "postgres",
						Status:     "running",
					}, nil).
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
			name: "rebalance maps to app error",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorByID(mock.Anything, "pg-connector").
					Return(domain.Connector{}, domain.ErrRebalanceInProgress).
					Once()
			},
			wantIs: errs.ErrRebalanceInProgress,
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
			name: "unexpected error passes through wrapped",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorByID(mock.Anything, "pg-connector").
					Return(domain.Connector{}, errSourceDown).
					Once()
			},
			wantIs: errSourceDown,
		},
		{
			name: "delete missing connector maps to app error",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorByID(mock.Anything, "pg-connector").
					Return(domain.Connector{Name: "pg-connector"}, nil).
					Once()

				kc.EXPECT().
					Delete(mock.Anything, "pg-connector").
					Return(domain.ErrConnectorNotFound).
					Once()
			},
			wantIs: errs.ErrConnectorNotFound,
		},
		{
			name: "delete error maps to app error",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorByID(mock.Anything, "pg-connector").
					Return(domain.Connector{Name: "pg-connector"}, nil).
					Once()

				kc.EXPECT().
					Delete(mock.Anything, "pg-connector").
					Return(domain.ErrRebalanceInProgress).
					Once()
			},
			wantIs: errs.ErrRebalanceInProgress,
		},
		{
			name: "delete unavailable maps to app error",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorByID(mock.Anything, "pg-connector").
					Return(domain.Connector{Name: "pg-connector"}, nil).
					Once()

				kc.EXPECT().
					Delete(mock.Anything, "pg-connector").
					Return(domain.ErrKafkaConnectUnavailable).
					Once()
			},
			wantIs: errs.ErrKafkaConnectUnavailable,
		},
		{
			name: "delete unexpected error passes through wrapped",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectorByID(mock.Anything, "pg-connector").
					Return(domain.Connector{Name: "pg-connector"}, nil).
					Once()

				kc.EXPECT().
					Delete(mock.Anything, "pg-connector").
					Return(errSourceDown).
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

			err := svc.Delete(testCtx(), "pg-connector")

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
		})
	}
}
