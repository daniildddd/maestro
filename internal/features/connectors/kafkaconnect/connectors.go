package kafkaconnect

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/daniildddd/maestro/internal/core/domain"
)

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
		return domain.Connector{}, fmt.Errorf("%s: %w", op, connectorError(err))
	}

	var status connectorStatus

	if err := c.do(ctx, http.MethodGet, path+"/status", nil, &status); err != nil {
		return domain.Connector{}, fmt.Errorf("%s: %w", op, connectorError(err))
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

func (c *HTTPClient) CreateConnector(ctx context.Context, name string, config map[string]string) (domain.Connector, error) {
	const op = "connectors.kafkaconnect.CreateConnector"

	body := struct {
		Name   string            `json:"name"`
		Config map[string]string `json:"config"`
	}{Name: name, Config: config}

	var created connectorInfo

	var err error

	for attempt := 1; attempt <= c.retryMaxAttempts; attempt++ {
		err = c.do(ctx, http.MethodPost, "/connectors", body, &created)
		if err == nil {
			break
		}

		if !isRebalance(err) || attempt == c.retryMaxAttempts {
			break
		}

		delay := c.retryWait(attempt)

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return domain.Connector{}, fmt.Errorf("%s: %w", op, ctx.Err())
		}
	}

	if err != nil {
		return domain.Connector{}, fmt.Errorf("%s: %w", op, connectorError(err))
	}

	pluginType := created.Config["connector.class"]

	connector, err := domain.NewConnector(
		created.Name,
		pluginType,
		c.sourceConfig(pluginType, created.Config),
		domain.ConnectorStatusStarting,
		make([]domain.Task, 0),
		0,
	)
	if err != nil {
		return domain.Connector{}, fmt.Errorf("%s: %w", op, err)
	}

	return connector, nil
}

func (c *HTTPClient) UpdateConnector(ctx context.Context, name string, config map[string]string) (domain.Connector, error) {
	const op = "connectors.kafkaconnect.UpdateConnector"

	var probe connectorInfo

	if err := c.do(ctx, http.MethodGet, "/connectors/"+url.PathEscape(name), nil, &probe); err != nil {
		return domain.Connector{}, fmt.Errorf("%s: %w", op, connectorError(err))
	}

	var updated connectorInfo

	var err error

	for attempt := 1; attempt <= c.retryMaxAttempts; attempt++ {
		err = c.do(ctx, http.MethodPut, "/connectors/"+url.PathEscape(name)+"/config", config, &updated)
		if err == nil {
			break
		}

		if !isRebalance(err) || attempt == c.retryMaxAttempts {
			break
		}

		delay := c.retryWait(attempt)

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return domain.Connector{}, fmt.Errorf("%s: %w", op, ctx.Err())
		}
	}

	if err != nil {
		return domain.Connector{}, fmt.Errorf("%s: %w", op, connectorError(err))
	}

	connector, err := c.GetConnectorByID(ctx, name)
	if err != nil {
		return domain.Connector{}, fmt.Errorf("%s: %w", op, err)
	}

	return connector, nil
}

func (c *HTTPClient) PauseConnector(ctx context.Context, name string) (domain.Connector, error) {
	return c.putConnectorAction(ctx, name, "pause")
}

func (c *HTTPClient) ResumeConnector(ctx context.Context, name string) (domain.Connector, error) {
	return c.putConnectorAction(ctx, name, "resume")
}

//nolint:revive // includeTasks/onlyFailed are restart options from the API contract
func (c *HTTPClient) RestartConnector(
	ctx context.Context,
	name string,
	includeTasks bool,
	onlyFailed bool,
) (domain.Connector, error) {
	const op = "connectors.kafkaconnect.RestartConnector"

	path := "/connectors/" + url.PathEscape(name) + "/restart"

	query := url.Values{}

	if includeTasks {
		query.Set("includeTasks", "true")
	}

	if onlyFailed {
		query.Set("onlyFailed", "true")
	}

	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var err error

	for attempt := 1; attempt <= c.retryMaxAttempts; attempt++ {
		err = c.do(ctx, http.MethodPost, path, nil, nil)
		if err == nil {
			break
		}

		if !isRebalance(err) || attempt == c.retryMaxAttempts {
			break
		}

		delay := c.retryWait(attempt)

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return domain.Connector{}, fmt.Errorf("%s: %w", op, ctx.Err())
		}
	}

	if err != nil {
		return domain.Connector{}, fmt.Errorf("%s: %w", op, connectorError(err))
	}

	connector, err := c.GetConnectorByID(ctx, name)
	if err != nil {
		return domain.Connector{}, fmt.Errorf("%s: %w", op, err)
	}

	return connector, nil
}

func (c *HTTPClient) Delete(ctx context.Context, name string) error {
	const op = "connectors.kafkaconnect.Delete"

	path := "/connectors/" + url.PathEscape(name)

	var err error

	for attempt := 1; attempt <= c.retryMaxAttempts; attempt++ {
		err = c.do(ctx, http.MethodDelete, path, nil, nil)
		if err == nil {
			return nil
		}

		if !isRebalance(err) || attempt == c.retryMaxAttempts {
			break
		}

		delay := c.retryWait(attempt)

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return fmt.Errorf("%s: %w", op, ctx.Err())
		}
	}

	return fmt.Errorf("%s: %w", op, connectorError(err))
}

func (c *HTTPClient) putConnectorAction(ctx context.Context, name, action string) (domain.Connector, error) {
	const op = "connectors.kafkaconnect.putConnectorAction"

	path := "/connectors/" + url.PathEscape(name) + "/" + action

	var err error

	for attempt := 1; attempt <= c.retryMaxAttempts; attempt++ {
		err = c.do(ctx, http.MethodPut, path, nil, nil)
		if err == nil {
			break
		}

		if !isRebalance(err) || attempt == c.retryMaxAttempts {
			break
		}

		delay := c.retryWait(attempt)

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return domain.Connector{}, fmt.Errorf("%s: %w", op, ctx.Err())
		}
	}

	if err != nil {
		return domain.Connector{}, fmt.Errorf("%s: %w", op, connectorError(err))
	}

	connector, err := c.GetConnectorByID(ctx, name)
	if err != nil {
		return domain.Connector{}, fmt.Errorf("%s: %w", op, err)
	}

	return connector, nil
}

func (c *HTTPClient) sourceConfig(pluginType string, config map[string]string) domain.SourceConfig {
	if adapter := c.registry.For(pluginType); adapter != nil {
		return adapter.Canonicalize(config)
	}

	return domain.SourceConfig{}
}
