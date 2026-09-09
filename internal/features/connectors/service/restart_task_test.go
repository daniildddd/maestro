package service_test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/features/connectors/service"
)

func TestRestartTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupMock func(kc *MockKafkaConnect)
		wantIs    error
	}{
		{
			name: "success restarts task",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					RestartTask(mock.Anything, "pg-connector", 0).
					Return(nil).
					Once()
			},
		},
		{
			name: "missing task maps to ErrConnectorTaskNotFound",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					RestartTask(mock.Anything, "pg-connector", 0).
					Return(domain.ErrConnectorTaskNotFound).
					Once()
			},
			wantIs: errs.ErrConnectorTaskNotFound,
		},
		{
			name: "rebalance maps to ErrRebalanceInProgress",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					RestartTask(mock.Anything, "pg-connector", 0).
					Return(domain.ErrRebalanceInProgress).
					Once()
			},
			wantIs: errs.ErrRebalanceInProgress,
		},
		{
			name: "connect unavailable maps to ErrKafkaConnectUnavailable",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					RestartTask(mock.Anything, "pg-connector", 0).
					Return(domain.ErrKafkaConnectUnavailable).
					Once()
			},
			wantIs: errs.ErrKafkaConnectUnavailable,
		},
		{
			name: "unexpected error passes through wrapped",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					RestartTask(mock.Anything, "pg-connector", 0).
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

			err := svc.RestartTask(testCtx(), "pg-connector", 0)

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
		})
	}
}
