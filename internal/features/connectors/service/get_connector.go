package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
)

func (s *ConnectorsService) GetConnectorByID(ctx context.Context, name string) (domain.Connector, error) {
	const op = "connectors.service.GetConnectorByID"

	connector, err := s.connectors.GetConnectorByID(ctx, name)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrConnectorNotFound):
			return domain.Connector{}, fmt.Errorf("%s: %w", op, errs.ErrConnectorNotFound)
		case errors.Is(err, domain.ErrKafkaConnectUnavailable):
			return domain.Connector{}, fmt.Errorf("%s: %w", op, errs.ErrKafkaConnectUnavailable)
		default:
			return domain.Connector{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	return connector, nil
}
