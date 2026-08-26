package kafkaconnect

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/features/connectors/plugins"
)

const (
	deleteMaxAttempts = 3
	deleteRetryDelay  = 300 * time.Millisecond
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

type connectorInfo struct {
	Name   string            `json:"name"`
	Config map[string]string `json:"config"`
}

type connectorTask struct {
	ID       int    `json:"id"`
	State    string `json:"state"`
	WorkerID string `json:"worker_id"`
}

type connectorStatus struct {
	Connector struct {
		State    string `json:"state"`
		WorkerID string `json:"worker_id"`
	} `json:"connector"`
	Tasks []connectorTask `json:"tasks"`
}

type listResponse map[string]connectorExpansion

type connectorExpansion struct {
	Status struct {
		Connector struct {
			State string `json:"state"`
		} `json:"connector"`
		Tasks []connectorTask `json:"tasks"`
	} `json:"status"`
	Info struct {
		Config map[string]string `json:"config"`
	} `json:"info"`
}

func (c *HTTPClient) GetConnectors(ctx context.Context) ([]domain.Connector, error) {
	const op = "connectors.kafkaconnect.GetConnectors"

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

func (c *HTTPClient) GetConnectorByID(ctx context.Context, name string) (domain.Connector, error) {
	const op = "connectors.kafkaconnect.GetConnectorByID"

	path := "/connectors/" + url.PathEscape(name)

	var info connectorInfo

	if err := c.do(ctx, http.MethodGet, path, nil, &info); err != nil {
		if isNotFound(err) {
			return domain.Connector{}, fmt.Errorf("%s: %w", op, domain.ErrConnectorNotFound)
		}

		return domain.Connector{}, fmt.Errorf("%s: %w", op, err)
	}

	var status connectorStatus

	if err := c.do(ctx, http.MethodGet, path+"/status", nil, &status); err != nil {
		if isNotFound(err) {
			return domain.Connector{}, fmt.Errorf("%s: %w", op, domain.ErrConnectorNotFound)
		}

		return domain.Connector{}, fmt.Errorf("%s: %w", op, err)
	}

	config := info.Config
	pluginType := config["connector.class"]

	tasks := make([]domain.Task, 0, len(status.Tasks))
	for _, t := range status.Tasks {
		tasks = append(tasks, domain.Task{ID: t.ID, State: strings.ToLower(t.State), WorkerID: t.WorkerID})
	}

	connector, err := domain.NewConnector(
		info.Name,
		pluginType,
		c.sourceConfig(pluginType, config),
		domain.DeriveConnectorStatus(status.Connector.State, len(tasks)),
		tasks,
		len(tasks),
	)
	if err != nil {
		return domain.Connector{}, fmt.Errorf("%s: %w", op, err)
	}

	connector.WorkerID = status.Connector.WorkerID

	return connector, nil
}

func (c *HTTPClient) Delete(ctx context.Context, name string) error {
	const op = "connectors.kafkaconnect.Delete"

	path := "/connectors/" + url.PathEscape(name)

	var err error

	for attempt := 1; attempt <= deleteMaxAttempts; attempt++ {
		err = c.do(ctx, http.MethodDelete, path, nil, nil)
		if err == nil {
			return nil
		}

		if !isRebalance(err) || attempt == deleteMaxAttempts {
			break
		}

		delay := deleteRetryDelay * time.Duration(attempt)

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return fmt.Errorf("%s: %w", op, ctx.Err())
		}
	}

	switch {
	case isNotFound(err):
		return fmt.Errorf("%s: %w", op, domain.ErrConnectorNotFound)
	case isRebalance(err):
		return fmt.Errorf("%s: %w", op, domain.ErrRebalanceInProgress)
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
}

func isNotFound(err error) bool {
	return strings.Contains(err.Error(), "unexpected status 404")
}

func isRebalance(err error) bool {
	return strings.Contains(err.Error(), "rebalance")
}

func (c *HTTPClient) sourceConfig(pluginType string, config map[string]string) domain.SourceConfig {
	if adapter := c.registry.For(pluginType); adapter != nil {
		return adapter.Canonicalize(config)
	}

	return domain.SourceConfig{}
}

//nolint:unparam // body будет использоваться запросами create/update
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

		return fmt.Errorf("%s: %w: %v", op, domain.ErrKafkaConnectUnavailable, err)
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
		return fmt.Errorf("%s: %w: unexpected status %d: %s", op, domain.ErrKafkaConnectUnavailable, resp.StatusCode, string(data))
	default:
		return fmt.Errorf("%s: unexpected status %d: %s", op, resp.StatusCode, string(data))
	}
}
