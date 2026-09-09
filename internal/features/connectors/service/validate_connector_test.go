package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/features/connectors/service"
)

const postgresClass = "io.debezium.connector.postgresql.PostgresConnector"

type checkerStub struct {
	check func(
		ctx context.Context,
		config map[string]string,
	) ([]domain.ValidationStep, error)
}

func (s checkerStub) Match(string) bool {
	return true
}

func (s checkerStub) Check(
	ctx context.Context,
	config map[string]string,
) ([]domain.ValidationStep, error) {
	return s.check(ctx, config)
}

func TestValidateConnector(t *testing.T) {
	t.Parallel()

	config := map[string]string{"database.hostname": "pg-1"}

	okConfigChecks := []domain.ValidationCheck{
		{
			ID:       "config.schema",
			Severity: domain.CheckSeverityOK,
		},
	}

	tests := []struct {
		name    string
		setupKC func(m *MockKafkaConnect)
		checker func() *checkerStub
		wantErr error
		want    domain.ValidationReport
	}{
		{
			name: "composes report from config and db steps",
			setupKC: func(m *MockKafkaConnect) {
				m.EXPECT().ValidateConfig(mock.Anything, postgresClass, config).
					Return(okConfigChecks, nil).
					Once()
			},
			checker: func() *checkerStub {
				return &checkerStub{check: func(
					_ context.Context,
					_ map[string]string,
				) ([]domain.ValidationStep, error) {
					return []domain.ValidationStep{{
						ID: domain.StepCDC,
						Checks: []domain.ValidationCheck{{
							ID:       "postgres.cdc.wal_level",
							Severity: domain.CheckSeverityError,
						}},
					}}, nil
				}}
			},
			want: domain.ValidationReport{
				Valid: false,
				Steps: []domain.ValidationStep{
					{
						ID:     domain.StepConfig,
						Status: domain.CheckSeverityOK,
						Checks: []domain.ValidationCheck{
							{
								ID:       "config.schema",
								Severity: domain.CheckSeverityOK,
							},
						},
					},
					{
						ID:     domain.StepCDC,
						Status: domain.CheckSeverityError,
						Checks: []domain.ValidationCheck{
							{
								ID:       "postgres.cdc.wal_level",
								Severity: domain.CheckSeverityError,
							},
						},
					},
				},
			},
		},
		{
			name: "config errors invalidate the report",
			setupKC: func(m *MockKafkaConnect) {
				m.EXPECT().ValidateConfig(mock.Anything, postgresClass, config).
					Return([]domain.ValidationCheck{{
						ID:       "config.database.hostname",
						Severity: domain.CheckSeverityError,
						Field:    "database.hostname",
					}}, nil).
					Once()
			},
			checker: func() *checkerStub {
				return &checkerStub{check: func(
					_ context.Context,
					_ map[string]string,
				) ([]domain.ValidationStep, error) {
					return nil, nil
				}}
			},
			want: domain.ValidationReport{
				Valid: false,
				Steps: []domain.ValidationStep{{
					ID:     domain.StepConfig,
					Status: domain.CheckSeverityError,
					Checks: []domain.ValidationCheck{{
						ID:       "config.database.hostname",
						Severity: domain.CheckSeverityError,
						Field:    "database.hostname",
					}},
				}},
			},
		},
		{
			name:    "unsupported connector returns 400 without calling Kafka Connect",
			setupKC: func(_ *MockKafkaConnect) {},
			checker: func() *checkerStub {
				return nil
			},
			wantErr: errs.ErrConnectorValidationUnsupported,
		},
		{
			name: "unknown plugin maps to plugin not found",
			setupKC: func(m *MockKafkaConnect) {
				m.EXPECT().ValidateConfig(mock.Anything, postgresClass, config).
					Return(nil, domain.ErrConnectorPluginNotFound).
					Once()
			},
			checker: func() *checkerStub {
				return &checkerStub{check: func(
					_ context.Context,
					_ map[string]string,
				) ([]domain.ValidationStep, error) {
					return nil, nil
				}}
			},
			wantErr: errs.ErrConnectorPluginNotFound,
		},
		{
			name: "kafka connect unavailable passes through unchanged",
			setupKC: func(m *MockKafkaConnect) {
				m.EXPECT().ValidateConfig(mock.Anything, postgresClass, config).
					Return(nil, domain.ErrKafkaConnectUnavailable).
					Once()
			},
			checker: func() *checkerStub {
				return &checkerStub{check: func(
					_ context.Context,
					_ map[string]string,
				) ([]domain.ValidationStep, error) {
					return nil, nil
				}}
			},
			wantErr: errs.ErrKafkaConnectUnavailable,
		},
		{
			name: "rebalance on config validation passes through wrapped",
			setupKC: func(m *MockKafkaConnect) {
				m.EXPECT().ValidateConfig(mock.Anything, postgresClass, config).
					Return(nil, domain.ErrRebalanceInProgress).
					Once()
			},
			checker: func() *checkerStub {
				return &checkerStub{check: func(
					_ context.Context,
					_ map[string]string,
				) ([]domain.ValidationStep, error) {
					return nil, nil
				}}
			},
			wantErr: domain.ErrRebalanceInProgress,
		},
		{
			name: "checker error is wrapped",
			setupKC: func(m *MockKafkaConnect) {
				m.EXPECT().ValidateConfig(mock.Anything, postgresClass, config).
					Return(okConfigChecks, nil).
					Once()
			},
			checker: func() *checkerStub {
				return &checkerStub{check: func(
					_ context.Context,
					_ map[string]string,
				) ([]domain.ValidationStep, error) {
					return nil, errSourceDown
				}}
			},
			wantErr: errSourceDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)

			kafkaConnect := NewMockKafkaConnect(t)
			tt.setupKC(kafkaConnect)

			var connectorsService *service.ConnectorsService

			if stub := tt.checker(); stub != nil {
				connectorsService = service.NewConnectorsService(kafkaConnect, noopAuditor{}, stub)
			} else {
				connectorsService = service.NewConnectorsService(kafkaConnect, noopAuditor{})
			}

			report, err := connectorsService.ValidateConnector(
				context.Background(), postgresClass, config,
			)

			if tt.wantErr != nil {
				must.ErrorIs(err, tt.wantErr)

				return
			}

			must.NoError(err)
			assert.Equal(t, tt.want, report)
		})
	}
}
