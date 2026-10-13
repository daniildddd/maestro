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

func TestConnectorsService_GetSMTPluginSchema(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "get smt plugin schema", func(a *gotest.Context) {
		allFields := []domain.ConnectorPluginField{
			{
				Name:       "transforms",
				Importance: domain.PluginImportanceHigh,
				Required:   true,
			},
			{
				Name:       "insert.field",
				Importance: domain.PluginImportanceMedium,
			},
			{
				Name:       "internal.option",
				Importance: domain.PluginImportanceLow,
			},
		}

		tests := []struct {
			name      string
			pluginID  string
			filter    *domain.ConnectorPluginSchemaFilter
			setupMock func(kc *MockKafkaConnect)
			want      domain.ConnectorPluginSchema
			wantIs    error
		}{
			{
				name:     "high filter keeps high fields",
				pluginID: "io.debezium.transforms.ExtractNewRecordState",
				filter:   mustSchemaFilter(t, "high"),
				setupMock: func(kc *MockKafkaConnect) {
					kc.EXPECT().
						GetConnectorPluginSchema(mock.Anything, "io.debezium.transforms.ExtractNewRecordState").
						Return(domain.ConnectorPluginSchema{Fields: allFields}, nil).
						Once()
				},
				want: domain.ConnectorPluginSchema{Fields: allFields[:1]},
			},
			{
				name:     "medium filter keeps high and medium fields",
				pluginID: "io.debezium.transforms.ExtractNewRecordState",
				filter:   mustSchemaFilter(t, "medium"),
				setupMock: func(kc *MockKafkaConnect) {
					kc.EXPECT().
						GetConnectorPluginSchema(mock.Anything, "io.debezium.transforms.ExtractNewRecordState").
						Return(domain.ConnectorPluginSchema{Fields: allFields}, nil).
						Once()
				},
				want: domain.ConnectorPluginSchema{Fields: allFields[:2]},
			},
			{
				name:     "all filter keeps everything",
				pluginID: "io.debezium.transforms.ExtractNewRecordState",
				filter:   mustSchemaFilter(t, "all"),
				setupMock: func(kc *MockKafkaConnect) {
					kc.EXPECT().
						GetConnectorPluginSchema(mock.Anything, "io.debezium.transforms.ExtractNewRecordState").
						Return(domain.ConnectorPluginSchema{Fields: allFields}, nil).
						Once()
				},
				want: domain.ConnectorPluginSchema{Fields: allFields},
			},
			{
				name:     "plugin not found maps to ErrSmtPluginNotFound",
				pluginID: "unknown.Transform",
				filter:   mustSchemaFilter(t, "high"),
				setupMock: func(kc *MockKafkaConnect) {
					kc.EXPECT().
						GetConnectorPluginSchema(mock.Anything, "unknown.Transform").
						Return(domain.ConnectorPluginSchema{}, domain.ErrConnectorPluginNotFound).
						Once()
				},
				wantIs: errs.ErrSmtPluginNotFound,
			},
			{
				name:     "connect unavailable maps to ErrKafkaConnectUnavailable",
				pluginID: "io.debezium.transforms.ExtractNewRecordState",
				filter:   mustSchemaFilter(t, "high"),
				setupMock: func(kc *MockKafkaConnect) {
					kc.EXPECT().
						GetConnectorPluginSchema(mock.Anything, "io.debezium.transforms.ExtractNewRecordState").
						Return(domain.ConnectorPluginSchema{}, domain.ErrKafkaConnectUnavailable).
						Once()
				},
				wantIs: errs.ErrKafkaConnectUnavailable,
			},
			{
				name:     "unexpected error passes through wrapped",
				pluginID: "io.debezium.transforms.ExtractNewRecordState",
				filter:   mustSchemaFilter(t, "high"),
				setupMock: func(kc *MockKafkaConnect) {
					kc.EXPECT().
						GetConnectorPluginSchema(mock.Anything, "io.debezium.transforms.ExtractNewRecordState").
						Return(domain.ConnectorPluginSchema{}, errSourceDown).
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

				got, err := svc.GetSMTPluginSchema(testCtx(), tt.pluginID, tt.filter)

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
		gotest.WithFeature("Get smt plugin schema"),
		gotest.WithSeverity("normal"),
		gotest.WithOwner("qa-intern"),
	)
}
