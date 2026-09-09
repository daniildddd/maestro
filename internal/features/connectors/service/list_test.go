package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/features/connectors/service"
)

var errSourceDown = errors.New("source down")

func TestConnectorsServiceGetConnectors(t *testing.T) {
	t.Parallel()

	pgConnector := domain.Connector{
		Name:       "pg-connector",
		PluginType: "io.debezium.connector.postgresql.PostgresConnector",
		Status:     domain.ConnectorStatusRunning,
		TasksCount: 1,
	}
	mysqlConnector := domain.Connector{
		Name:       "mysql-connector",
		PluginType: "io.debezium.connector.mysql.MySqlConnector",
		Status:     domain.ConnectorStatusPaused,
		TasksCount: 2,
	}
	exoticConnector := domain.Connector{
		Name:       "exotic",
		PluginType: "io.debezium.connector.vitess.VitessConnector",
		Status:     domain.ConnectorStatusRunning,
	}

	tests := []struct {
		name      string
		filter    *domain.ConnectorFilter
		setupMock func(kc *MockKafkaConnect)
		want      []domain.Connector
		wantIs    error
	}{
		{
			name:   "returns connectors as-is",
			filter: mustFilter(t, 1, 20, "", ""),
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectors(mock.Anything).
					Return([]domain.Connector{pgConnector, exoticConnector}, nil).
					Once()
			},
			want: []domain.Connector{pgConnector, exoticConnector},
		},
		{
			name:   "status filter keeps only matching connectors",
			filter: mustFilter(t, 1, 20, domain.ConnectorStatusPaused, ""),
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectors(mock.Anything).
					Return([]domain.Connector{pgConnector, mysqlConnector}, nil).
					Once()
			},
			want: []domain.Connector{mysqlConnector},
		},
		{
			name:   "search matches name substring case-insensitively",
			filter: mustFilter(t, 1, 20, "", "MYSQL"),
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectors(mock.Anything).
					Return([]domain.Connector{pgConnector, mysqlConnector}, nil).
					Once()
			},
			want: []domain.Connector{mysqlConnector},
		},
		{
			name:   "pagination slices filtered result",
			filter: mustFilter(t, 2, 1, "", ""),
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectors(mock.Anything).
					Return([]domain.Connector{pgConnector, mysqlConnector}, nil).
					Once()
			},
			want: []domain.Connector{mysqlConnector},
		},
		{
			name:   "pagination beyond end returns empty list",
			filter: mustFilter(t, 10, 20, "", ""),
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectors(mock.Anything).
					Return([]domain.Connector{pgConnector}, nil).
					Once()
			},
			want: []domain.Connector{},
		},
		{
			name:   "kafka connect unavailable maps to ErrKafkaConnectUnavailable",
			filter: mustFilter(t, 1, 20, "", ""),
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectors(mock.Anything).
					Return(nil, domain.ErrKafkaConnectUnavailable).
					Once()
			},
			wantIs: errs.ErrKafkaConnectUnavailable,
		},
		{
			name:   "unexpected error passes through",
			filter: mustFilter(t, 1, 20, "", ""),
			setupMock: func(kc *MockKafkaConnect) {
				kc.EXPECT().
					GetConnectors(mock.Anything).
					Return(nil, errSourceDown).
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

			got, err := svc.GetConnectors(context.Background(), tt.filter)

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			must.Equal(tt.want, got)
		})
	}
}

func mustFilter(t *testing.T, page, limit int, status, search string) *domain.ConnectorFilter {
	t.Helper()

	filter, err := domain.NewConnectorFilter(page, limit, status, search)
	if err != nil {
		t.Fatalf("build filter: %v", err)
	}

	return filter
}
