package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
)

func (s *ConnectorsService) RestartTask(
	ctx context.Context,
	connectorName string,
	taskID int,
) error {
	const op = "connectors.service.RestartTask"

	if err := s.connectors.RestartTask(ctx, connectorName, taskID); err != nil {
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

	actorID := reqctx.UserID(ctx)

	s.auditor.Record(ctx, domain.AuditEvent{
		ID:          uuid.New(),
		Action:      domain.ActionConnectorTaskRestarted,
		Outcome:     domain.OutcomeSuccess,
		ActorID:     actorID,
		SubjectType: domain.AuditSubjectConnector,
		SubjectID:   connectorName,
		SubjectName: connectorName,
		StateAfter: map[string]any{
			"task_id": taskID,
		},
		RequestID: reqctx.RequestID(ctx).String(),
		IP:        reqctx.ClientIP(ctx),
		UserAgent: reqctx.UserAgent(ctx),
		CreatedAt: time.Now(),
	})

	return nil
}
