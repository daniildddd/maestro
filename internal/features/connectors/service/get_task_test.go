package service_test

import (
	"testing"

	gotest "github.com/allure-framework/allure-go/commons/gotest" //nolint:revive // package clause is allure, gotest keeps call sites readable
	allurerequire "github.com/allure-framework/allure-go/testify/require"
	"github.com/stretchr/testify/mock"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/features/connectors/service"
)

func TestConnectorsService_GetTaskByID(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "get connector task", func(a *gotest.Context) {
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
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

				kc := NewMockKafkaConnect(a.T())
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
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Get connector task"),
		gotest.WithSeverity("normal"),
		gotest.WithOwner("qa-intern"),
	)
}
