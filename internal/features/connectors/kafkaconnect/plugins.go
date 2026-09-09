package kafkaconnect

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/daniildddd/maestro/internal/core/domain"
)

type pluginInfo struct {
	Class   string `json:"class"`
	Type    string `json:"type"`
	Version string `json:"version"`
}

func (c *HTTPClient) GetConnectorPlugins(ctx context.Context) ([]domain.ConnectorPlugin, error) {
	const op = "connectors.kafkaconnect.GetConnectorPlugins"

	infos, err := c.getPlugins(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	conPlugins := make([]domain.ConnectorPlugin, 0, len(infos))

	for _, p := range infos {
		conPlugins = append(conPlugins, domain.ConnectorPlugin{ID: p.Class})
	}

	return conPlugins, nil
}

func (c *HTTPClient) GetSMTPlugins(ctx context.Context) ([]domain.ConnectorPlugin, error) {
	const op = "connectors.kafkaconnect.GetSMTPlugins"

	infos, err := c.getPlugins(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	plugins := make([]domain.ConnectorPlugin, 0, len(infos))

	for _, p := range infos {
		if p.Type != "transformation" {
			continue
		}

		plugins = append(plugins, domain.ConnectorPlugin{ID: p.Class})
	}

	return plugins, nil
}

type pluginConfigKey struct {
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Required      bool    `json:"required"`
	DefaultValue  *string `json:"default_value"`
	Importance    string  `json:"importance"`
	Documentation *string `json:"documentation"`
	DisplayName   *string `json:"display_name"`
}

func (c *HTTPClient) GetConnectorPluginSchema(ctx context.Context, pluginID string) (domain.ConnectorPluginSchema, error) {
	const op = "connectors.kafkaconnect.GetConnectorPluginSchema"

	path := "/connector-plugins/" + url.PathEscape(pluginID) + "/config"

	var resp []pluginConfigKey

	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		if isNotFound(err) {
			return domain.ConnectorPluginSchema{}, fmt.Errorf("%s: %w", op, domain.ErrConnectorPluginNotFound)
		}

		return domain.ConnectorPluginSchema{}, fmt.Errorf("%s: %w", op, err)
	}

	fields := make([]domain.ConnectorPluginField, 0, len(resp))

	for _, key := range resp {
		fields = append(fields, domain.ConnectorPluginField{
			Name:        key.Name,
			Label:       key.DisplayName,
			Description: key.Documentation,
			Type:        key.Type,
			Importance:  key.Importance,
			Required:    key.Required,
			Default:     key.DefaultValue,
		})
	}

	return domain.ConnectorPluginSchema{Fields: fields}, nil
}

func (c *HTTPClient) GetConnectorPluginSchemaWithValues(ctx context.Context, pluginID string) (domain.ConnectorPluginSchema, error) {
	const op = "connectors.kafkaconnect.GetConnectorPluginSchemaWithValues"

	schema, err := c.GetConnectorPluginSchema(ctx, pluginID)
	if err != nil {
		return domain.ConnectorPluginSchema{}, fmt.Errorf("%s: %w", op, err)
	}

	values, err := c.pluginRecommendedValues(ctx, pluginID)
	if err != nil {
		return domain.ConnectorPluginSchema{}, fmt.Errorf("%s: %w", op, err)
	}

	for i := range schema.Fields {
		if v, ok := values[schema.Fields[i].Name]; ok {
			schema.Fields[i].Values = v
		}
	}

	return schema, nil
}

type validateConfigResponse struct {
	Configs []struct {
		Value struct {
			Name   string   `json:"name"`
			Errors []string `json:"errors"`
		} `json:"value"`
	} `json:"configs"`
}

func (c *HTTPClient) ValidateConfig(
	ctx context.Context,
	pluginID string,
	config map[string]string,
) ([]domain.ValidationCheck, error) {
	const op = "connectors.kafkaconnect.ValidateConfig"

	var resp validateConfigResponse

	if err := c.do(ctx, http.MethodPut,
		"/connector-plugins/"+url.PathEscape(pluginID)+"/config/validate",
		config, &resp); err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrConnectorPluginNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	checks := make([]domain.ValidationCheck, 0, len(resp.Configs))

	for _, cfg := range resp.Configs {
		if len(cfg.Value.Errors) == 0 {
			continue
		}

		message := strings.Join(cfg.Value.Errors, "; ")
		if message == "" {
			message = "invalid value"
		}

		checks = append(checks, domain.ValidationCheck{
			ID:       "config." + cfg.Value.Name,
			Severity: domain.CheckSeverityError,
			Message:  message,
			Field:    cfg.Value.Name,
		})
	}

	if len(checks) == 0 {
		checks = append(checks, domain.ValidationCheck{
			ID:       "config.schema",
			Severity: domain.CheckSeverityOK,
			Message:  "configuration accepted by Kafka Connect",
		})
	}

	return checks, nil
}

//nolint:revive // connectorsOnly mirrors the kafka connect query parameter of the same name
func (c *HTTPClient) getPlugins(ctx context.Context, connectorsOnly bool) ([]pluginInfo, error) {
	const op = "connectors.kafkaconnect.getPlugins"

	path := "/connector-plugins"

	if !connectorsOnly {
		path += "?connectorsOnly=false"
	}

	var resp []pluginInfo

	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *HTTPClient) pluginRecommendedValues(ctx context.Context, pluginID string) (map[string][]string, error) {
	const op = "connectors.kafkaconnect.pluginRecommendedValues"

	var infos struct {
		Configs []struct {
			Value struct {
				Name              string   `json:"name"`
				RecommendedValues []string `json:"recommended_values"`
			} `json:"value"`
		} `json:"configs"`
	}

	body := map[string]string{"connector.class": pluginID}

	if err := c.do(ctx, http.MethodPut,
		"/connector-plugins/"+url.PathEscape(pluginID)+"/config/validate",
		body, &infos); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	values := make(map[string][]string)

	for _, cfg := range infos.Configs {
		if len(cfg.Value.RecommendedValues) == 0 {
			continue
		}

		values[cfg.Value.Name] = uniqueStrings(cfg.Value.RecommendedValues)
	}

	return values, nil
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))

	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}

		seen[s] = struct{}{}
		out = append(out, s)
	}

	return out
}
