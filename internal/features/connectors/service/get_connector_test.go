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

func TestConnectorsService_GetConnectorByID(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "get connector by id", func(a *gotest.Context) {
		tests := []struct {
			name      string
			setupMock func(kc *MockKafkaConnect)
			want      domain.Connector
			wantIs    error
		}{
			{
				name: "success returns connector",
				setupMock: func(kc *MockKafkaConnect) {
					kc.EXPECT().
						GetConnectorByID(mock.Anything, "pg-connector").
						Return(domain.Connector{
							Name:       "pg-connector",
							PluginType: "io.debezium.connector.postgresql.PostgresConnector",
							Status:     domain.ConnectorStatusRunning,
						}, nil).
						Once()
				},
				want: domain.Connector{
					Name:       "pg-connector",
					PluginType: "io.debezium.connector.postgresql.PostgresConnector",
					Status:     domain.ConnectorStatusRunning,
				},
			},
			{
				name: "missing connector maps to ErrConnectorNotFound",
				setupMock: func(kc *MockKafkaConnect) {
					kc.EXPECT().
						GetConnectorByID(mock.Anything, "pg-connector").
						Return(domain.Connector{}, domain.ErrConnectorNotFound).
						Once()
				},
				wantIs: errs.ErrConnectorNotFound,
			},
			{
				name: "connect unavailable maps to ErrKafkaConnectUnavailable",
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
		}

		for _, tt := range tests {
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

				kc := NewMockKafkaConnect(a.T())
				tt.setupMock(kc)

				svc := service.NewConnectorsService(kc, noopAuditor{})

				connector, err := svc.GetConnectorByID(testCtx(), "pg-connector")

				if tt.wantIs != nil {
					must.ErrorIs(err, tt.wantIs)

					return
				}

				must.NoError(err)
				must.Equal(tt.want, connector)
			})
		}
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Get connector"),
		gotest.WithSeverity("normal"),
		gotest.WithOwner("qa-intern"),
	)
}
