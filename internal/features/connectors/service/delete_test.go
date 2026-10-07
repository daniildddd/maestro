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

func TestConnectorsService_Delete(t *testing.T) { //nolint:paralleltest // allure gotest.Test owns subtest execution
	gotest.Test(t, "delete connector", func(a *gotest.Context) {
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
				name: "get missing connector maps to ErrConnectorNotFound",
				setupMock: func(kc *MockKafkaConnect) {
					kc.EXPECT().
						GetConnectorByID(mock.Anything, "pg-connector").
						Return(domain.Connector{}, domain.ErrConnectorNotFound).
						Once()
				},
				wantIs: errs.ErrConnectorNotFound,
			},
			{
				name: "delete missing connector maps to ErrConnectorNotFound",
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
				name: "delete error maps to ErrRebalanceInProgress",
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
				name: "delete unavailable maps to ErrKafkaConnectUnavailable",
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
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

				kc := NewMockKafkaConnect(a.T())
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

		a.Step("records connector state in audit", func(a *gotest.Context) {
			must := allurerequire.New(a)

			dbName := "shop"
			pluginName := "pgoutput"

			kc := NewMockKafkaConnect(a.T())
			kc.EXPECT().
				GetConnectorByID(mock.Anything, "pg-connector").
				Return(domain.Connector{
					Name:       "pg-connector",
					PluginType: "postgres",
					Status:     "running",
					Config: domain.SourceConfig{
						Hostname:   "pg-1",
						Port:       "5432",
						User:       "debezium",
						DBName:     &dbName,
						PluginName: &pluginName,
					},
				}, nil).
				Once()

			kc.EXPECT().
				Delete(mock.Anything, "pg-connector").
				Return(nil).
				Once()

			auditor := &capturingAuditor{}
			svc := service.NewConnectorsService(kc, auditor)

			err := svc.Delete(testCtx(), "pg-connector")
			must.NoError(err)
			must.Len(auditor.recorded, 1)

			state := auditor.recorded[0].StateBefore

			must.Equal(map[string]any{
				"name":              "pg-connector",
				"plugin_type":       "postgres",
				"status":            "running",
				"database.hostname": "pg-1",
				"database.port":     "5432",
				"database.user":     "debezium",
				"database.dbname":   "shop",
				"plugin.name":       "pgoutput",
			}, state)
		})
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Delete connector"),
		gotest.WithSeverity("critical"),
		gotest.WithOwner("qa-intern"),
	)
}
