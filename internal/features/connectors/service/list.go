package service

import (
	"context"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func (s *ConnectorsService) GetConnectors(
	ctx context.Context,
	filter *domain.ConnectorFilter,
) ([]domain.Connector, error) {
	const op = "connectors.service.GetConnectors"

	connectors, err := s.connectors.GetConnectors(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	matched := filter.Apply(connectors)

	return paginateConnectors(matched, filter), nil
}

func paginateConnectors(connectors []domain.Connector, filter *domain.ConnectorFilter) []domain.Connector {
	offset := (filter.Page - 1) * filter.Limit

	if offset >= len(connectors) {
		return []domain.Connector{}
	}

	end := min(offset+filter.Limit, len(connectors))

	return connectors[offset:end]
}
