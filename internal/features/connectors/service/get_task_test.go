package service_test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/features/connectors/service"
)

func TestGetTaskByID(t *testing.T) {
	t.Parallel()

	task := domain.Task{
		ID:       0,
		State:    "RUNNING",
		WorkerID: "worker-1",
	}

	tests := []struct {
		name      string
		setupMock func(kc *MockKafkaConnect)
		want      domain.Task
		wantIs    error
	}{
		{
			name: "success returns task",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetTaskByID(mock.Anything, "pg-connector", 0).
					Return(task, nil).
					Once()
			},
			want: task,
		},
		{
			name: "missing task maps to ErrConnectorTaskNotFound",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetTaskByID(mock.Anything, "pg-connector", 0).
					Return(domain.Task{}, domain.ErrConnectorTaskNotFound).
					Once()
			},
			wantIs: errs.ErrConnectorTaskNotFound,
		},
		{
			name: "connect unavailable maps to ErrKafkaConnectUnavailable",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetTaskByID(mock.Anything, "pg-connector", 0).
					Return(domain.Task{}, domain.ErrKafkaConnectUnavailable).
					Once()
			},
			wantIs: errs.ErrKafkaConnectUnavailable,
		},
		{
			name: "unexpected error passes through wrapped",
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetTaskByID(mock.Anything, "pg-connector", 0).
					Return(domain.Task{}, errSourceDown).
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

			got, err := svc.GetTaskByID(testCtx(), "pg-connector", 0)

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			must.Equal(tt.want, got)
		})
	}
}
