package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
)

func (s *ConnectorsService) GetSMTPluginSchema(
	ctx context.Context,
	pluginID string,
	filter *domain.ConnectorPluginSchemaFilter,
) (domain.ConnectorPluginSchema, error) {
	const op = "connectors.service.GetSMTPluginSchema"

	schema, err := s.connectors.GetConnectorPluginSchema(ctx, pluginID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrConnectorPluginNotFound):
			return domain.ConnectorPluginSchema{}, fmt.Errorf("%s: %w", op, errs.ErrSmtPluginNotFound)
		case errors.Is(err, domain.ErrKafkaConnectUnavailable):
			return domain.ConnectorPluginSchema{}, fmt.Errorf("%s: %w", op, errs.ErrKafkaConnectUnavailable)
		default:
			return domain.ConnectorPluginSchema{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	schema.Fields = filter.Apply(schema.Fields)

	return schema, nil
}
