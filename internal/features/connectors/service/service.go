package service

import (
	"context"

	"github.com/daniildddd/maestro/internal/core/domain"
)

type dbChecker interface {
	Match(class string) bool

	Check(
		ctx context.Context,
		config map[string]string,
	) ([]domain.ValidationStep, error)
}

type ConnectorsService struct {
	connectors KafkaConnect
	checkers   []dbChecker
}

func NewConnectorsService(
	connectors KafkaConnect,
	checkers ...dbChecker,
) *ConnectorsService {
	return &ConnectorsService{
		connectors: connectors,
		checkers:   checkers,
	}
}

//nolint:interfacebloat // the kafka connect client contract grows with connector features
type KafkaConnect interface {
	GetConnectors(
		ctx context.Context,
	) ([]domain.Connector, error)

	GetConnectorByID(
		ctx context.Context,
		name string,
	) (domain.Connector, error)

	GetConnectorPlugins(
		ctx context.Context,
	) ([]domain.ConnectorPlugin, error)

	GetConnectorPluginSchema(
		ctx context.Context,
		pluginID string,
	) (domain.ConnectorPluginSchema, error)

	GetConnectorPluginSchemaWithValues(
		ctx context.Context,
		pluginID string,
	) (domain.ConnectorPluginSchema, error)

	GetSMTPlugins(
		ctx context.Context,
	) ([]domain.ConnectorPlugin, error)

	CreateConnector(
		ctx context.Context,
		name string,
		config map[string]string,
	) (domain.Connector, error)

	GetTaskByID(
		ctx context.Context,
		name string,
		taskID int,
	) (domain.Task, error)

	RestartTask(
		ctx context.Context,
		name string,
		taskID int,
	) error

	UpdateConnector(
		ctx context.Context,
		name string,
		config map[string]string,
	) (domain.Connector, error)

	PauseConnector(
		ctx context.Context,
		name string,
	) (domain.Connector, error)

	ResumeConnector(
		ctx context.Context,
		name string,
	) (domain.Connector, error)

	RestartConnector(
		ctx context.Context,
		name string,
		includeTasks bool,
		onlyFailed bool,
	) (domain.Connector, error)

	Delete(
		ctx context.Context,
		name string,
	) error

	ValidateConfig(
		ctx context.Context,
		pluginID string,
		config map[string]string,
	) ([]domain.ValidationCheck, error)
}
