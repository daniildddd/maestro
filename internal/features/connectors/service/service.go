package service

import (
	"context"

	"github.com/daniildddd/maestro/internal/core/domain"
)

type ConnectorsService struct {
	connectors KafkaConnect
}

func NewConnectorsService(
	connectors KafkaConnect,
) *ConnectorsService {
	return &ConnectorsService{
		connectors: connectors,
	}
}

type KafkaConnect interface {
	GetConnectors(
		ctx context.Context,
	) ([]domain.Connector, error)

	GetConnectorByID(
		ctx context.Context,
		name string,
	) (domain.Connector, error)

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
}
