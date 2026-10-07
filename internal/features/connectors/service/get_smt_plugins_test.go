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

func TestConnectorsService_GetSMTPlugins(t *testing.T) { //nolint:paralleltest // allure gotest.Test owns subtest execution
	gotest.Test(t, "get smt plugins", func(a *gotest.Context) {
		plugins := []domain.ConnectorPlugin{
			{
				ID: "io.debezium.transforms.ExtractNewRecordState",
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
						GetSMTPlugins(mock.Anything).
						Return(plugins, nil).
						Once()
				},
				want: plugins,
			},
			{
				name: "connect unavailable maps to ErrKafkaConnectUnavailable",
				setupMock: func(kc *MockKafkaConnect) {
					kc.EXPECT().
						GetSMTPlugins(mock.Anything).
						Return(nil, domain.ErrKafkaConnectUnavailable).
						Once()
				},
				wantIs: errs.ErrKafkaConnectUnavailable,
			},
			{
				name: "unexpected error passes through wrapped",
				setupMock: func(kc *MockKafkaConnect) {
					kc.EXPECT().
						GetSMTPlugins(mock.Anything).
						Return(nil, errSourceDown).
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

				got, err := svc.GetSMTPlugins(testCtx())

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
		gotest.WithFeature("Get smt plugins"),
		gotest.WithSeverity("normal"),
		gotest.WithOwner("qa-intern"),
	)
}
