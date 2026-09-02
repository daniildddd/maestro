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

func (s *ConnectorsService) CreateConnector(
	ctx context.Context,
	name string,
	config map[string]string,
) (domain.Connector, error) {
	const op = "connectors.service.CreateConnector"

	connector, err := s.connectors.CreateConnector(ctx, name, config)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrConnectorAlreadyExists):
			return domain.Connector{}, fmt.Errorf("%s: %w", op, errs.ErrConnectorAlreadyExists)
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
		Action:      domain.ActionConnectorCreated,
		Outcome:     domain.OutcomeSuccess,
		ActorID:     actorID,
		SubjectType: domain.AuditSubjectConnector,
		SubjectID:   connector.Name,
		SubjectName: connector.Name,
		StateAfter:  s.maskedConfig(ctx, config),
		RequestID:   reqctx.RequestID(ctx).String(),
		IP:          reqctx.ClientIP(ctx),
		UserAgent:   reqctx.UserAgent(ctx),
		CreatedAt:   time.Now(),
	})

	return connector, nil
}

func (s *ConnectorsService) maskedConfig(
	ctx context.Context,
	config map[string]string,
) map[string]any {
	secretKeys, ok := s.secretKeys(ctx, config["connector.class"])
	if !ok {
		return nil
	}

	masked := make(map[string]any, len(config))

	for key, value := range config {
		if _, isSecret := secretKeys[key]; isSecret {
			masked[key] = "***"

			continue
		}

		masked[key] = value
	}

	return masked
}

func (s *ConnectorsService) secretKeys(
	ctx context.Context,
	pluginID string,
) (map[string]struct{}, bool) {
	if pluginID == "" {
		return nil, false
	}

	schema, err := s.connectors.GetConnectorPluginSchema(ctx, pluginID)
	if err != nil {
		return nil, false
	}

	keys := make(map[string]struct{}, len(schema.Fields))

	for _, field := range schema.Fields {
		if field.Type == "PASSWORD" {
			keys[field.Name] = struct{}{}
		}
	}

	return keys, true
}
