package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
		StateAfter:  maskedConfig(config),
		RequestID:   reqctx.RequestID(ctx).String(),
		IP:          reqctx.ClientIP(ctx),
		UserAgent:   reqctx.UserAgent(ctx),
		CreatedAt:   time.Now(),
	})

	return connector, nil
}

func maskedConfig(config map[string]string) map[string]any {
	masked := make(map[string]any, len(config))

	for key, value := range config {
		if isSecretKey(key) {
			masked[key] = "***"

			continue
		}

		masked[key] = value
	}

	return masked
}

func isSecretKey(key string) bool {
	lower := strings.ToLower(key)

	return strings.Contains(lower, "password") ||
		strings.Contains(lower, "secret") ||
		strings.Contains(lower, "token")
}
