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
	List(
		ctx context.Context,
	) ([]domain.Connector, error)
}
