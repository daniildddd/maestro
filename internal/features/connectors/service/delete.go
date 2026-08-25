package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
)

func (s *ConnectorsService) Delete(ctx context.Context, name string) error {
	const op = "connectors.service.Delete"

	if err := s.connectors.Delete(ctx, name); err != nil {
		switch {
		case errors.Is(err, domain.ErrConnectorNotFound):
			return fmt.Errorf("%s: %w", op, errs.ErrConnectorNotFound)
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
