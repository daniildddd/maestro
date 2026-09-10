package service_test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/features/connectors/service"
)

func TestConnectorsService_GetSMTPluginSchema(t *testing.T) {
	t.Parallel()

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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			kc := NewMockKafkaConnect(t)
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
}
