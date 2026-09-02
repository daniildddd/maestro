package service

import (
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
