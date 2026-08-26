package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
)

func (s *ConnectorsService) CreateConnector(
	ctx context.Context,
	name string,
	config map[string]string,
) (domain.Connector, error) {
	const op = "connectors.service.CreateConnector"

	connector, err := s.connectors.CreateConnector(ctx, name, config)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrConnectorAlreadyExists):
			return domain.Connector{}, fmt.Errorf("%s: %w", op, errs.ErrConnectorAlreadyExists)
		case errors.Is(err, domain.ErrRebalanceInProgress):
			return domain.Connector{}, fmt.Errorf("%s: %w", op, errs.ErrRebalanceInProgress)
		case errors.Is(err, domain.ErrInvalidConnectorConfig):
			return domain.Connector{}, fmt.Errorf("%s: %w", op, errs.ErrValidationFailed)
		case errors.Is(err, domain.ErrKafkaConnectUnavailable):
			return domain.Connector{}, fmt.Errorf("%s: %w", op, errs.ErrKafkaConnectUnavailable)
		default:
			return domain.Connector{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	return connector, nil
}
