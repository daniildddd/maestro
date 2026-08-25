package domain

import (
	"errors"
	"fmt"
	"strings"
)

const (
	ConnectorStatusRunning  = "running"
	ConnectorStatusPaused   = "paused"
	ConnectorStatusFailed   = "failed"
	ConnectorStatusStarting = "starting"
)

var (
	ErrInvalidConnector       = errors.New("invalid connector")
	ErrInvalidConnectorStatus = errors.New("status must be one of: running, paused, failed, starting")
	ErrConnectorNotFound       = errors.New("connector not found")
	ErrKafkaConnectUnavailable = errors.New("kafka connect unavailable")
	ErrRebalanceInProgress     = errors.New("kafka connect rebalance in progress")
)

type Task struct {
	ID       int
	State    string
	WorkerID string
}

type SourceConfig struct {
	Hostname   string
	Port       string
	User       string
	DBName     *string
	PluginName *string
}

type Connector struct {
	Name       string
	PluginType string
	Config     SourceConfig
	Status     string
	TasksCount int
	Tasks      []Task
}

func NewConnector(
	name string,
	pluginType string,
	config SourceConfig,
	status string,
	tasks []Task,
	tasksCount int,
) (Connector, error) {
	const op = "core.domain.NewConnector"

	c := Connector{
		Name:       name,
		PluginType: pluginType,
		Config:     config,
		Status:     status,
		Tasks:      tasks,
		TasksCount: tasksCount,
	}

	if err := c.Validate(); err != nil {
		return Connector{}, fmt.Errorf("%s: %w", op, err)
	}

	return c, nil
}

func DeriveConnectorStatus(connectorState string, tasksCount int) string {
	switch strings.ToLower(connectorState) {
	case ConnectorStatusRunning:
		if tasksCount > 0 {
			return ConnectorStatusRunning
		}
	case ConnectorStatusPaused:
		return ConnectorStatusPaused
	case ConnectorStatusFailed:
		return ConnectorStatusFailed
	default:
	}

	return ConnectorStatusStarting
}

func (c Connector) Validate() error {
	const op = "core.domain.Connector.Validate"

	if c.Name == "" {
		return fmt.Errorf("%s: %w: name is required", op, ErrInvalidConnector)
	}

	if c.PluginType == "" {
		return fmt.Errorf("%s: %w: plugin_type is required", op, ErrInvalidConnector)
	}

	if c.Status != ConnectorStatusRunning &&
		c.Status != ConnectorStatusPaused &&
		c.Status != ConnectorStatusFailed &&
		c.Status != ConnectorStatusStarting {
		return fmt.Errorf("%s: %w: %q", op, ErrInvalidConnectorStatus, c.Status)
	}

	if c.TasksCount != len(c.Tasks) {
		return fmt.Errorf("%s: %w: tasks_count %d != len(tasks) %d", op, ErrInvalidConnector, c.TasksCount, len(c.Tasks))
	}

	return nil
}

type ConnectorFilter struct {
	Page   int
	Limit  int
	Status string
	Search string
}

func NewConnectorFilter(
	page int,
	limit int,
	status string,
	search string,
) (*ConnectorFilter, error) {
	const op = "core.domain.NewConnectorFilter"

	f := &ConnectorFilter{
		Page:   page,
		Limit:  limit,
		Status: status,
		Search: search,
	}
	f.Normalize()

	if err := f.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return f, nil
}

func (f *ConnectorFilter) Normalize() {
	if f.Page <= 0 {
		f.Page = 1
	}

	if f.Limit <= 0 {
		f.Limit = 20
	}

	if f.Limit > 100 {
		f.Limit = 100
	}
}

func (f *ConnectorFilter) Validate() error {
	const op = "core.domain.ConnectorFilter.Validate"

	if f.Status != "" &&
		f.Status != ConnectorStatusRunning &&
		f.Status != ConnectorStatusPaused &&
		f.Status != ConnectorStatusFailed &&
		f.Status != ConnectorStatusStarting {
		return fmt.Errorf("%s: %w", op, ErrInvalidConnectorStatus)
	}

	return nil
}

func (f *ConnectorFilter) Apply(connectors []Connector) []Connector {
	matched := make([]Connector, 0)

	for _, connector := range connectors {
		if f.Status != "" && connector.Status != f.Status {
			continue
		}

		if f.Search != "" && !strings.Contains(strings.ToLower(connector.Name), strings.ToLower(f.Search)) {
			continue
		}

		matched = append(matched, connector)
	}

	return matched
}
