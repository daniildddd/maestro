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

func (s *ConnectorsService) UpdateConnector(
	ctx context.Context,
	name string,
	config map[string]string,
) (domain.Connector, error) {
	const op = "connectors.service.UpdateConnector"

	before, err := s.connectors.GetConnectorByID(ctx, name)
	if err != nil {
		return domain.Connector{}, fmt.Errorf("%s: get connector: %w", op, mapConnectorError(err))
	}

	connector, err := s.connectors.UpdateConnector(ctx, name, config)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrConnectorNotFound):
			return domain.Connector{}, fmt.Errorf("%s: %w", op, errs.ErrConnectorNotFound)
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

	actorID := reqctx.UserID(ctx)

	s.auditor.Record(ctx, domain.AuditEvent{
		ID:          uuid.New(),
		Action:      domain.ActionConnectorUpdated,
		Outcome:     domain.OutcomeSuccess,
		ActorID:     actorID,
		ActorLogin:  reqctx.Username(ctx),
		SubjectType: domain.AuditSubjectConnector,
		SubjectID:   connector.Name,
		SubjectName: connector.Name,
		StateBefore: connectorState(before),
		StateAfter:  s.maskedConfig(ctx, config),
		RequestID:   reqctx.RequestID(ctx).String(),
		IP:          reqctx.ClientIP(ctx),
		UserAgent:   reqctx.UserAgent(ctx),
		CreatedAt:   time.Now(),
	})

	return connector, nil
}
