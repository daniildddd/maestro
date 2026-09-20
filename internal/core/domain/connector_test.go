package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func TestNewConnector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		connName   string
		pluginType string
		config     domain.SourceConfig
		status     string
		tasks      []domain.Task
		tasksCount int
		want       domain.Connector
		wantErr    error
	}{
		{
			name:       "valid running connector accepted",
			connName:   "orders-pg",
			pluginType: "PostgresSource",
			config: domain.SourceConfig{
				Hostname:   "localhost",
				Port:       "5432",
				User:       "user",
				DBName:     new("db"),
				PluginName: new("pgoutput"),
			},
			status: "running",
			tasks: []domain.Task{
				{ID: 0, State: "running", WorkerID: "worker-1", Trace: new("trace")},
			},
			tasksCount: 1,
			want: domain.Connector{
				Name:       "orders-pg",
				PluginType: "PostgresSource",
				Config: domain.SourceConfig{
					Hostname:   "localhost",
					Port:       "5432",
					User:       "user",
					DBName:     new("db"),
					PluginName: new("pgoutput"),
				},
				Status: "running",
				Tasks: []domain.Task{
					{ID: 0, State: "running", WorkerID: "worker-1", Trace: new("trace")},
				},
				TasksCount: 1,
			},
		},
		{
			name:       "valid starting connector without tasks accepted",
			connName:   "orders-pg",
			pluginType: "PostgresSource",
			config: domain.SourceConfig{
				Hostname: "localhost",
				Port:     "5432",
				User:     "user",
			},
			status:     "starting",
			tasks:      nil,
			tasksCount: 0,
			want: domain.Connector{
				Name:       "orders-pg",
				PluginType: "PostgresSource",
				Config: domain.SourceConfig{
					Hostname: "localhost",
					Port:     "5432",
					User:     "user",
				},
				Status:     "starting",
				Tasks:      nil,
				TasksCount: 0,
			},
		},
		{
			name:       "valid paused connector accepted",
			connName:   "orders-pg",
			pluginType: "PostgresSource",
			config: domain.SourceConfig{
				Hostname: "localhost",
				Port:     "5432",
				User:     "user",
			},
			status:     "paused",
			tasks:      []domain.Task{},
			tasksCount: 0,
			want: domain.Connector{
				Name:       "orders-pg",
				PluginType: "PostgresSource",
				Config: domain.SourceConfig{
					Hostname: "localhost",
					Port:     "5432",
					User:     "user",
				},
				Status:     "paused",
				Tasks:      []domain.Task{},
				TasksCount: 0,
			},
		},
		{
			name:       "empty name rejected",
			connName:   "",
			pluginType: "PostgresSource",
			config: domain.SourceConfig{
				Hostname: "localhost",
				Port:     "5432",
				User:     "user",
			},
			status:     "running",
			tasks:      nil,
			tasksCount: 0,
			wantErr:    domain.ErrInvalidConnector,
		},
		{
			name:       "empty plugin type rejected",
			connName:   "orders-pg",
			pluginType: "",
			config: domain.SourceConfig{
				Hostname: "localhost",
				Port:     "5432",
				User:     "user",
			},
			status:     "running",
			tasks:      nil,
			tasksCount: 0,
			wantErr:    domain.ErrInvalidConnector,
		},
		{
			name:       "unknown status rejected",
			connName:   "orders-pg",
			pluginType: "PostgresSource",
			config: domain.SourceConfig{
				Hostname: "localhost",
				Port:     "5432",
				User:     "user",
			},
			status:     "unknown",
			tasks:      nil,
			tasksCount: 0,
			wantErr:    domain.ErrInvalidConnectorStatus,
		},
		{
			name:       "empty status rejected",
			connName:   "orders-pg",
			pluginType: "PostgresSource",
			config: domain.SourceConfig{
				Hostname: "localhost",
				Port:     "5432",
				User:     "user",
			},
			status:     "",
			tasks:      nil,
			tasksCount: 0,
			wantErr:    domain.ErrInvalidConnectorStatus,
		},
		{
			name:       "tasks count mismatch rejected",
			connName:   "orders-pg",
			pluginType: "PostgresSource",
			config: domain.SourceConfig{
				Hostname: "localhost",
				Port:     "5432",
				User:     "user",
			},
			status: "running",
			tasks: []domain.Task{
				{ID: 0, State: "running", WorkerID: "worker-1"},
			},
			tasksCount: 2,
			wantErr:    domain.ErrInvalidConnector,
		},
		{
			name:       "nil tasks with nonzero count rejected",
			connName:   "orders-pg",
			pluginType: "PostgresSource",
			config: domain.SourceConfig{
				Hostname: "localhost",
				Port:     "5432",
				User:     "user",
			},
			status:     "running",
			tasks:      nil,
			tasksCount: 1,
			wantErr:    domain.ErrInvalidConnector,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			must := require.New(t)

			got, err := domain.NewConnector(tt.connName, tt.pluginType, tt.config, tt.status, tt.tasks, tt.tasksCount)

			if tt.wantErr != nil {
				must.ErrorIs(err, tt.wantErr)
				is.Equal(domain.Connector{}, got)

				return
			}

			must.NoError(err)
			is.Equal(tt.want, got)
		})
	}
}

func TestDeriveConnectorStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		state      string
		tasksCount int
		want       string
	}{
		{
			name:       "running with tasks stays running",
			state:      "running",
			tasksCount: 2,
			want:       "running",
		},
		{
			name:       "running without tasks becomes starting",
			state:      "running",
			tasksCount: 0,
			want:       "starting",
		},
		{
			name:       "paused with tasks stays paused",
			state:      "paused",
			tasksCount: 2,
			want:       "paused",
		},
		{
			name:       "paused without tasks stays paused",
			state:      "paused",
			tasksCount: 0,
			want:       "paused",
		},
		{
			name:       "failed with tasks stays failed",
			state:      "failed",
			tasksCount: 2,
			want:       "failed",
		},
		{
			name:       "failed without tasks stays failed",
			state:      "failed",
			tasksCount: 0,
			want:       "failed",
		},
		{
			name:       "starting stays starting",
			state:      "starting",
			tasksCount: 2,
			want:       "starting",
		},
		{
			name:       "unknown state becomes starting",
			state:      "unassigned",
			tasksCount: 2,
			want:       "starting",
		},
		{
			name:       "empty state becomes starting",
			state:      "",
			tasksCount: 0,
			want:       "starting",
		},
		{
			name:       "uppercase running with tasks stays running",
			state:      "RUNNING",
			tasksCount: 1,
			want:       "running",
		},
		{
			name:       "uppercase running without tasks becomes starting",
			state:      "RUNNING",
			tasksCount: 0,
			want:       "starting",
		},
		{
			name:       "mixed case paused stays paused",
			state:      "Paused",
			tasksCount: 0,
			want:       "paused",
		},
		{
			name:       "mixed case failed stays failed",
			state:      "Failed",
			tasksCount: 1,
			want:       "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)

			got := domain.DeriveConnectorStatus(tt.state, tt.tasksCount)

			is.Equal(tt.want, got)
		})
	}
}

func TestConnector_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		connector domain.Connector
		wantErr   error
	}{
		{
			name: "valid running connector accepted",
			connector: domain.Connector{
				Name:       "orders-pg",
				PluginType: "PostgresSource",
				Config: domain.SourceConfig{
					Hostname: "localhost",
					Port:     "5432",
					User:     "user",
				},
				Status: "running",
				Tasks: []domain.Task{
					{ID: 0, State: "running", WorkerID: "worker-1"},
				},
				TasksCount: 1,
			},
		},
		{
			name: "valid starting connector without tasks accepted",
			connector: domain.Connector{
				Name:       "orders-pg",
				PluginType: "PostgresSource",
				Status:     "starting",
				Tasks:      nil,
				TasksCount: 0,
			},
		},
		{
			name: "empty name rejected",
			connector: domain.Connector{
				Name:       "",
				PluginType: "PostgresSource",
				Status:     "running",
				Tasks:      nil,
				TasksCount: 0,
			},
			wantErr: domain.ErrInvalidConnector,
		},
		{
			name: "empty plugin type rejected",
			connector: domain.Connector{
				Name:       "orders-pg",
				PluginType: "",
				Status:     "running",
				Tasks:      nil,
				TasksCount: 0,
			},
			wantErr: domain.ErrInvalidConnector,
		},
		{
			name: "unknown status rejected",
			connector: domain.Connector{
				Name:       "orders-pg",
				PluginType: "PostgresSource",
				Status:     "restarting",
				Tasks:      nil,
				TasksCount: 0,
			},
			wantErr: domain.ErrInvalidConnectorStatus,
		},
		{
			name: "empty status rejected",
			connector: domain.Connector{
				Name:       "orders-pg",
				PluginType: "PostgresSource",
				Status:     "",
				Tasks:      nil,
				TasksCount: 0,
			},
			wantErr: domain.ErrInvalidConnectorStatus,
		},
		{
			name: "tasks count greater than tasks rejected",
			connector: domain.Connector{
				Name:       "orders-pg",
				PluginType: "PostgresSource",
				Status:     "running",
				Tasks: []domain.Task{
					{ID: 0, State: "running", WorkerID: "worker-1"},
				},
				TasksCount: 2,
			},
			wantErr: domain.ErrInvalidConnector,
		},
		{
			name: "tasks count smaller than tasks rejected",
			connector: domain.Connector{
				Name:       "orders-pg",
				PluginType: "PostgresSource",
				Status:     "running",
				Tasks: []domain.Task{
					{ID: 0, State: "running", WorkerID: "worker-1"},
					{ID: 1, State: "running", WorkerID: "worker-1"},
				},
				TasksCount: 1,
			},
			wantErr: domain.ErrInvalidConnector,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			err := tt.connector.Validate()

			if tt.wantErr != nil {
				must.ErrorIs(err, tt.wantErr)

				return
			}

			must.NoError(err)
		})
	}
}

func TestNewConnectorFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		page    int
		limit   int
		status  string
		search  string
		want    domain.ConnectorFilter
		wantErr error
	}{
		{
			name:   "valid filter accepted",
			page:   1,
			limit:  20,
			status: "running",
			search: "orders",
			want: domain.ConnectorFilter{
				Page:   1,
				Limit:  20,
				Status: "running",
				Search: "orders",
			},
		},
		{
			name:   "empty status and search accepted",
			page:   2,
			limit:  10,
			status: "",
			search: "",
			want: domain.ConnectorFilter{
				Page:   2,
				Limit:  10,
				Status: "",
				Search: "",
			},
		},
		{
			name:   "zero page and limit get defaults",
			page:   0,
			limit:  0,
			status: "",
			search: "",
			want: domain.ConnectorFilter{
				Page:   1,
				Limit:  20,
				Status: "",
				Search: "",
			},
		},
		{
			name:   "negative page normalized",
			page:   -5,
			limit:  10,
			status: "",
			search: "",
			want: domain.ConnectorFilter{
				Page:   1,
				Limit:  10,
				Status: "",
				Search: "",
			},
		},
		{
			name:   "negative limit normalized",
			page:   1,
			limit:  -5,
			status: "",
			search: "",
			want: domain.ConnectorFilter{
				Page:   1,
				Limit:  20,
				Status: "",
				Search: "",
			},
		},
		{
			name:   "limit capped at 100",
			page:   1,
			limit:  500,
			status: "",
			search: "",
			want: domain.ConnectorFilter{
				Page:   1,
				Limit:  100,
				Status: "",
				Search: "",
			},
		},
		{
			name:    "unknown status rejected",
			page:    1,
			limit:   20,
			status:  "restarting",
			search:  "",
			wantErr: domain.ErrInvalidConnectorStatus,
		},
		{
			name:    "uppercase status rejected",
			page:    1,
			limit:   20,
			status:  "Running",
			search:  "",
			wantErr: domain.ErrInvalidConnectorStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			must := require.New(t)

			got, err := domain.NewConnectorFilter(tt.page, tt.limit, tt.status, tt.search)

			if tt.wantErr != nil {
				must.ErrorIs(err, tt.wantErr)
				must.Nil(got)

				return
			}

			must.NoError(err)
			is.Equal(tt.want, *got)
		})
	}
}

func TestConnectorFilter_Normalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input domain.ConnectorFilter
		want  domain.ConnectorFilter
	}{
		{
			name:  "valid values kept",
			input: domain.ConnectorFilter{Page: 2, Limit: 10, Status: "running", Search: "orders"},
			want:  domain.ConnectorFilter{Page: 2, Limit: 10, Status: "running", Search: "orders"},
		},
		{
			name:  "zero page and limit get defaults",
			input: domain.ConnectorFilter{Page: 0, Limit: 0},
			want:  domain.ConnectorFilter{Page: 1, Limit: 20},
		},
		{
			name:  "negative page normalized",
			input: domain.ConnectorFilter{Page: -3, Limit: 10},
			want:  domain.ConnectorFilter{Page: 1, Limit: 10},
		},
		{
			name:  "negative limit normalized",
			input: domain.ConnectorFilter{Page: 1, Limit: -7},
			want:  domain.ConnectorFilter{Page: 1, Limit: 20},
		},
		{
			name:  "limit capped at 100",
			input: domain.ConnectorFilter{Page: 1, Limit: 500},
			want:  domain.ConnectorFilter{Page: 1, Limit: 100},
		},
		{
			name:  "limit exactly 100 kept",
			input: domain.ConnectorFilter{Page: 1, Limit: 100},
			want:  domain.ConnectorFilter{Page: 1, Limit: 100},
		},
		{
			name:  "status and search preserved during normalization",
			input: domain.ConnectorFilter{Page: 0, Limit: 500, Status: "paused", Search: "pg"},
			want:  domain.ConnectorFilter{Page: 1, Limit: 100, Status: "paused", Search: "pg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)

			got := tt.input
			got.Normalize()

			is.Equal(tt.want, got)
		})
	}
}

func TestConnectorFilter_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		filter  domain.ConnectorFilter
		wantErr error
	}{
		{
			name:   "empty status accepted",
			filter: domain.ConnectorFilter{Page: 1, Limit: 20},
		},
		{
			name:   "running status accepted",
			filter: domain.ConnectorFilter{Page: 1, Limit: 20, Status: "running"},
		},
		{
			name:   "paused status accepted",
			filter: domain.ConnectorFilter{Page: 1, Limit: 20, Status: "paused"},
		},
		{
			name:   "failed status accepted",
			filter: domain.ConnectorFilter{Page: 1, Limit: 20, Status: "failed"},
		},
		{
			name:   "starting status accepted",
			filter: domain.ConnectorFilter{Page: 1, Limit: 20, Status: "starting"},
		},
		{
			name:    "unknown status rejected",
			filter:  domain.ConnectorFilter{Page: 1, Limit: 20, Status: "restarting"},
			wantErr: domain.ErrInvalidConnectorStatus,
		},
		{
			name:    "uppercase status rejected",
			filter:  domain.ConnectorFilter{Page: 1, Limit: 20, Status: "Running"},
			wantErr: domain.ErrInvalidConnectorStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			filter := tt.filter
			err := filter.Validate()

			if tt.wantErr != nil {
				must.ErrorIs(err, tt.wantErr)

				return
			}

			must.NoError(err)
		})
	}
}

func TestConnectorFilter_Apply(t *testing.T) {
	t.Parallel()

	connectors := []domain.Connector{
		{Name: "orders-pg", PluginType: "PostgresSource", Status: "running"},
		{Name: "orders-archive", PluginType: "PostgresSource", Status: "paused"},
		{Name: "payments-pg", PluginType: "PostgresSource", Status: "running"},
		{Name: "inventory", PluginType: "PostgresSource", Status: "failed"},
	}

	tests := []struct {
		name   string
		filter domain.ConnectorFilter
		want   []domain.Connector
	}{
		{
			name:   "empty filter returns all",
			filter: domain.ConnectorFilter{Page: 1, Limit: 20},
			want: []domain.Connector{
				{Name: "orders-pg", PluginType: "PostgresSource", Status: "running"},
				{Name: "orders-archive", PluginType: "PostgresSource", Status: "paused"},
				{Name: "payments-pg", PluginType: "PostgresSource", Status: "running"},
				{Name: "inventory", PluginType: "PostgresSource", Status: "failed"},
			},
		},
		{
			name:   "status filter returns matching connectors",
			filter: domain.ConnectorFilter{Page: 1, Limit: 20, Status: "running"},
			want: []domain.Connector{
				{Name: "orders-pg", PluginType: "PostgresSource", Status: "running"},
				{Name: "payments-pg", PluginType: "PostgresSource", Status: "running"},
			},
		},
		{
			name:   "search filter returns matching connectors",
			filter: domain.ConnectorFilter{Page: 1, Limit: 20, Search: "orders"},
			want: []domain.Connector{
				{Name: "orders-pg", PluginType: "PostgresSource", Status: "running"},
				{Name: "orders-archive", PluginType: "PostgresSource", Status: "paused"},
			},
		},
		{
			name:   "search is case insensitive",
			filter: domain.ConnectorFilter{Page: 1, Limit: 20, Search: "ORDERS"},
			want: []domain.Connector{
				{Name: "orders-pg", PluginType: "PostgresSource", Status: "running"},
				{Name: "orders-archive", PluginType: "PostgresSource", Status: "paused"},
			},
		},
		{
			name:   "status and search combined",
			filter: domain.ConnectorFilter{Page: 1, Limit: 20, Status: "running", Search: "orders"},
			want: []domain.Connector{
				{Name: "orders-pg", PluginType: "PostgresSource", Status: "running"},
			},
		},
		{
			name:   "no matches return empty",
			filter: domain.ConnectorFilter{Page: 1, Limit: 20, Status: "failed", Search: "orders"},
			want:   []domain.Connector{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)

			filter := tt.filter
			got := filter.Apply(connectors)

			is.Equal(tt.want, got)
		})
	}
}
