package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
)

func (s *ConnectorsService) ResumeConnector(ctx context.Context, name string) (domain.Connector, error) {
	const op = "connectors.service.ResumeConnector"

	connector, err := s.connectors.ResumeConnector(ctx, name)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrConnectorNotFound):
			return domain.Connector{}, fmt.Errorf("%s: %w", op, errs.ErrConnectorNotFound)
		case errors.Is(err, domain.ErrRebalanceInProgress):
			return domain.Connector{}, fmt.Errorf("%s: %w", op, errs.ErrRebalanceInProgress)
		case errors.Is(err, domain.ErrKafkaConnectUnavailable):
			return domain.Connector{}, fmt.Errorf("%s: %w", op, errs.ErrKafkaConnectUnavailable)
		default:
			return domain.Connector{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	return connector, nil
}
