package kafkaconnect

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/features/connectors/plugins"
)

type HTTPClient struct {
	baseURL  string
	client   *http.Client
	registry *plugins.Registry
}

func NewHTTPClient(cfg Config, registry *plugins.Registry) *HTTPClient {
	return &HTTPClient{
		baseURL:  cfg.BaseURL,
		client:   &http.Client{Timeout: cfg.Timeout},
		registry: registry,
	}
}

type listResponse map[string]connectorExpansion

type connectorExpansion struct {
	Status struct {
		Connector struct {
			State string `json:"state"`
		} `json:"connector"`
		Tasks []struct {
			ID       int    `json:"id"`
			State    string `json:"state"`
			WorkerID string `json:"worker_id"`
		} `json:"tasks"`
	} `json:"status"`
	Info struct {
		Config map[string]string `json:"config"`
	} `json:"info"`
}

func (c *HTTPClient) List(ctx context.Context) ([]domain.Connector, error) {
	const op = "connectors.kafkaconnect.List"

	var resp listResponse

	if err := c.do(ctx, http.MethodGet, "/connectors?expand=status&expand=info", nil, &resp); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	connectors := make([]domain.Connector, 0, len(resp))

	for name, v := range resp {
		config := v.Info.Config
		pluginType := config["connector.class"]
		status := domain.DeriveConnectorStatus(v.Status.Connector.State, len(v.Status.Tasks))

		tasks := make([]domain.Task, 0, len(v.Status.Tasks))

		for _, t := range v.Status.Tasks {
			tasks = append(tasks, domain.Task{ID: t.ID, State: strings.ToLower(t.State), WorkerID: t.WorkerID})
		}

		connector, err := domain.NewConnector(
			name,
			pluginType,
			c.sourceConfig(pluginType, config),
			status,
			tasks,
			len(tasks),
		)
		if err != nil {
			return nil, fmt.Errorf("%s: build connector %q: %w", op, name, err)
		}

		connectors = append(connectors, connector)
	}

	return connectors, nil
}

func (c *HTTPClient) sourceConfig(pluginType string, config map[string]string) domain.SourceConfig {
	if adapter := c.registry.For(pluginType); adapter != nil {
		return adapter.Canonicalize(config)
	}

	return domain.SourceConfig{}
}

func (c *HTTPClient) do(ctx context.Context, method, path string, body, out any) error {
	const op = "connectors.kafkaconnect.do"

	var reader io.Reader

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("%s: marshal body: %w", op, err)
		}

		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("%s: build request: %w", op, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return fmt.Errorf("%s: %w", op, err)
		}

		return fmt.Errorf("%s: %w: %v", op, errs.ErrKafkaConnectUnavailable, err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // body fully read below; close error not actionable

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s: read body: %w", op, err)
	}

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		if out != nil && len(data) > 0 {
			if err = json.Unmarshal(data, out); err != nil {
				return fmt.Errorf("%s: unmarshal response: %w", op, err)
			}
		}

		return nil
	case resp.StatusCode >= 500:
		return fmt.Errorf("%s: %w: unexpected status %d: %s", op, errs.ErrKafkaConnectUnavailable, resp.StatusCode, string(data))
	default:
		return fmt.Errorf("%s: unexpected status %d: %s", op, resp.StatusCode, string(data))
	}
}
