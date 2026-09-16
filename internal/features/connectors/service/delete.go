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

func (s *ConnectorsService) Delete(ctx context.Context, name string) error {
	const op = "connectors.service.Delete"

	connector, err := s.connectors.GetConnectorByID(ctx, name)
	if err != nil {
		return fmt.Errorf("%s: get connector: %w", op, mapConnectorError(err))
	}

	if err := s.connectors.Delete(ctx, name); err != nil {
		return fmt.Errorf("%s: %w", op, mapConnectorError(err))
	}

	actorID := reqctx.UserID(ctx)

	s.auditor.Record(ctx, domain.AuditEvent{
		ID:          uuid.New(),
		Action:      domain.ActionConnectorDeleted,
		Outcome:     domain.OutcomeSuccess,
		ActorID:     actorID,
		ActorLogin:  reqctx.Username(ctx),
		SubjectType: domain.AuditSubjectConnector,
		SubjectID:   name,
		SubjectName: name,
		StateBefore: connectorState(connector),
		RequestID:   reqctx.RequestID(ctx).String(),
		IP:          reqctx.ClientIP(ctx),
		UserAgent:   reqctx.UserAgent(ctx),
		CreatedAt:   time.Now(),
	})

	return nil
}

func mapConnectorError(err error) error {
	switch {
	case errors.Is(err, domain.ErrConnectorNotFound):
		return errs.ErrConnectorNotFound
	case errors.Is(err, domain.ErrRebalanceInProgress):
		return errs.ErrRebalanceInProgress
	case errors.Is(err, domain.ErrKafkaConnectUnavailable):
		return errs.ErrKafkaConnectUnavailable
	default:
		return err
	}
}
