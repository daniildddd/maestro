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

func (s *ConnectorsService) RestartConnector(
	ctx context.Context,
	name string,
	includeTasks bool,
	onlyFailed bool,
) (domain.Connector, error) {
	const op = "connectors.service.RestartConnector"

	connector, err := s.connectors.RestartConnector(ctx, name, includeTasks, onlyFailed)
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

	actorID := reqctx.UserID(ctx)

	s.auditor.Record(ctx, domain.AuditEvent{
		ID:          uuid.New(),
		Action:      domain.ActionConnectorRestarted,
		Outcome:     domain.OutcomeSuccess,
		ActorID:     actorID,
		SubjectType: domain.AuditSubjectConnector,
		SubjectID:   connector.Name,
		SubjectName: connector.Name,
		StateAfter:  connectorState(connector),
		RequestID:   reqctx.RequestID(ctx).String(),
		IP:          reqctx.ClientIP(ctx),
		UserAgent:   reqctx.UserAgent(ctx),
		CreatedAt:   time.Now(),
	})

	return connector, nil
}
