package service

import (
	"context"

	"github.com/daniildddd/maestro/internal/core/domain"
)

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
