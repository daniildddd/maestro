package service_test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/features/connectors/service"
)

func TestConnectorsService_Delete(t *testing.T) {
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

	t.Run("records connector state in audit", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		dbName := "shop"
		pluginName := "pgoutput"

		kc := NewMockKafkaConnect(t)
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
}
