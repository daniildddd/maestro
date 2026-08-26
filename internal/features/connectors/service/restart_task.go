package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
)

func (s *ConnectorsService) RestartTask(ctx context.Context, name string, taskID int) error {
	const op = "connectors.service.RestartTask"

	if err := s.connectors.RestartTask(ctx, name, taskID); err != nil {
		switch {
		case errors.Is(err, domain.ErrConnectorTaskNotFound):
			return fmt.Errorf("%s: %w", op, errs.ErrConnectorTaskNotFound)
		case errors.Is(err, domain.ErrRebalanceInProgress):
			return fmt.Errorf("%s: %w", op, errs.ErrRebalanceInProgress)
		case errors.Is(err, domain.ErrKafkaConnectUnavailable):
			return fmt.Errorf("%s: %w", op, errs.ErrKafkaConnectUnavailable)
		default:
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}
