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

func connectorState(connector domain.Connector) map[string]any {
	state := map[string]any{
		"name":        connector.Name,
		"plugin_type": connector.PluginType,
		"status":      connector.Status,
	}

	if connector.Config.Hostname != "" {
		state["database.hostname"] = connector.Config.Hostname
	}

	if connector.Config.Port != "" {
		state["database.port"] = connector.Config.Port
	}

	if connector.Config.User != "" {
		state["database.user"] = connector.Config.User
	}

	if connector.Config.DBName != nil {
		state["database.dbname"] = *connector.Config.DBName
	}

	if connector.Config.PluginName != nil {
		state["plugin.name"] = *connector.Config.PluginName
	}

	return state
}
