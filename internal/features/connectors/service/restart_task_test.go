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

func TestConnectorsService_RestartTask(t *testing.T) { //nolint:paralleltest // allure gotest.Test owns subtest execution
	gotest.Test(t, "restart connector task", func(a *gotest.Context) {
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
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

				kc := NewMockKafkaConnect(a.T())
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
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Restart connector task"),
		gotest.WithSeverity("critical"),
		gotest.WithOwner("qa-intern"),
	)
}
