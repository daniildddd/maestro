package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
)

func (s *ConnectorsService) GetTaskByID(ctx context.Context, name string, taskID int) (domain.Task, error) {
	const op = "connectors.service.GetTaskByID"

	task, err := s.connectors.GetTaskByID(ctx, name, taskID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrConnectorTaskNotFound):
			return domain.Task{}, fmt.Errorf("%s: %w", op, errs.ErrConnectorTaskNotFound)
		case errors.Is(err, domain.ErrKafkaConnectUnavailable):
			return domain.Task{}, fmt.Errorf("%s: %w", op, errs.ErrKafkaConnectUnavailable)
		default:
			return domain.Task{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	return task, nil
}
