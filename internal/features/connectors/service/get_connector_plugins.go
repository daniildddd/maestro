package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
)

func (s *ConnectorsService) GetConnectorPlugins(ctx context.Context) ([]domain.ConnectorPlugin, error) {
	const op = "connectors.service.GetConnectorPlugins"

	plugins, err := s.connectors.GetConnectorPlugins(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrKafkaConnectUnavailable) {
			return nil, fmt.Errorf("%s: %w", op, errs.ErrKafkaConnectUnavailable)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return plugins, nil
}
