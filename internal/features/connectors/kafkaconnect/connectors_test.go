package kafkaconnect_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func TestHTTPClient_GetConnectors(t *testing.T) {
	t.Parallel()

	t.Run("list returns connectors of any engine without failures", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			response := listResponse{
				"pg-connector": newExpansion(
					"io.debezium.connector.postgresql.PostgresConnector",
					"RUNNING",
					map[string]string{
						"database.hostname": "maestro-postgres",
						"database.port":     "5432",
						"database.user":     "postgres",
						"database.dbname":   "testdb",
						"plugin.name":       "pgoutput",
					},
				),
				"mysql-connector": newExpansion(
					"io.debezium.connector.mysql.MySqlConnector",
					"RUNNING",
					map[string]string{
						"database.hostname":     "mysql",
						"database.port":         "3306",
						"database.user":         "debezium",
						"database.include.list": "inventory,orders",
					},
				),
				"mongo-connector": newExpansion(
					"io.debezium.connector.mongodb.MongoDbConnector",
					"PAUSED",
					map[string]string{
						"mongodb.hosts": "mongo:27017",
						"mongodb.name":  "inventory",
					},
				),
				"unknown-connector": newExpansion(
					"io.debezium.connector.vitess.VitessConnector",
					"FAILED",
					map[string]string{
						"key.converter": "org.apache.kafka.connect.json.JsonConverter",
					},
				),
			}

			writeJSON(t, w, response)
		})

		connectors, err := client.GetConnectors(context.Background())

		must.NoError(err)
		must.ElementsMatch(
			[]domain.Connector{
				{
					Name:       "pg-connector",
					PluginType: "io.debezium.connector.postgresql.PostgresConnector",
					Status:     domain.ConnectorStatusRunning,
					TasksCount: 1,
					Tasks: []domain.Task{
						{ID: 0, State: "running", WorkerID: "worker-1"},
					},
					Config: domain.SourceConfig{
						Hostname:   "maestro-postgres",
						Port:       "5432",
						User:       "postgres",
						DBName:     strPtr("testdb"),
						PluginName: strPtr("pgoutput"),
					},
				},
				{
					Name:       "mysql-connector",
					PluginType: "io.debezium.connector.mysql.MySqlConnector",
					Status:     domain.ConnectorStatusRunning,
					TasksCount: 1,
					Tasks: []domain.Task{
						{ID: 0, State: "running", WorkerID: "worker-1"},
					},
					Config: domain.SourceConfig{
						Hostname: "mysql",
						Port:     "3306",
						User:     "debezium",
					},
				},
				{
					Name:       "mongo-connector",
					PluginType: "io.debezium.connector.mongodb.MongoDbConnector",
					Status:     domain.ConnectorStatusPaused,
					TasksCount: 1,
					Tasks: []domain.Task{
						{ID: 0, State: "paused", WorkerID: "worker-1"},
					},
					Config: domain.SourceConfig{},
				},
				{
					Name:       "unknown-connector",
					PluginType: "io.debezium.connector.vitess.VitessConnector",
					Status:     domain.ConnectorStatusFailed,
					TasksCount: 1,
					Tasks: []domain.Task{
						{ID: 0, State: "failed", WorkerID: "worker-1"},
					},
					Config: domain.SourceConfig{},
				},
			},
			connectors,
		)
	})

	t.Run("invalid entry fails with invalid connector", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(t, w, listResponse{
				"broken": newExpansion("", "RUNNING", map[string]string{
					"key.converter": "org.apache.kafka.connect.json.JsonConverter",
				}),
			})
		})

		_, err := client.GetConnectors(context.Background())

		must.ErrorIs(err, domain.ErrInvalidConnector)
	})

	t.Run("server error maps to kafka connect unavailable", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		_, err := client.GetConnectors(context.Background())

		must.ErrorIs(err, domain.ErrKafkaConnectUnavailable)
	})

	t.Run("canceled context returns context canceled", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		client := newTestClient(t, func(_ http.ResponseWriter, r *http.Request) {
			cancel()

			select {
			case <-r.Context().Done():
			case <-time.After(2 * time.Second):
				t.Error("request context was not canceled")
			}
		})

		_, err := client.GetConnectors(ctx)

		must.ErrorIs(err, context.Canceled)
		must.NotErrorIs(err, domain.ErrKafkaConnectUnavailable)
	})
}

func TestHTTPClient_GetConnectorByID(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/connectors/pg-connector":
			writeJSON(t, w, map[string]any{
				"name": "pg-connector",
				"config": map[string]string{
					"connector.class":   "io.debezium.connector.postgresql.PostgresConnector",
					"database.hostname": "pg-1",
					"database.dbname":   "testdb",
				},
			})
		case "/connectors/pg-connector/status":
			writeJSON(t, w, map[string]any{
				"connector": map[string]any{"state": "RUNNING", "worker_id": "worker-1"},
				"tasks": []map[string]any{
					{"id": 0, "state": "RUNNING", "worker_id": "worker-1"},
				},
			})
		default:
			http.NotFound(w, nil)
		}
	}))
	t.Cleanup(srv.Close)

	client := newTestClientOnServer(t, srv)

	connector, err := client.GetConnectorByID(context.Background(), "pg-connector")

	must.NoError(err)
	must.Equal(
		domain.Connector{
			Name:       "pg-connector",
			PluginType: "io.debezium.connector.postgresql.PostgresConnector",
			Status:     domain.ConnectorStatusRunning,
			WorkerID:   "worker-1",
			TasksCount: 1,
			Tasks: []domain.Task{
				{ID: 0, State: "running", WorkerID: "worker-1"},
			},
			Config: domain.SourceConfig{
				Hostname: "pg-1",
				DBName:   strPtr("testdb"),
			},
		},
		connector,
	)
}

func TestHTTPClient_CreateConnector(t *testing.T) {
	t.Parallel()

	create := func(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) (domain.Connector, error) {
		t.Helper()

		srv := httptest.NewServer(http.HandlerFunc(handler))
		t.Cleanup(srv.Close)

		client := newTestClientOnServer(t, srv)

		return client.CreateConnector(context.Background(), "pg-connector", map[string]string{
			"connector.class":   "io.debezium.connector.postgresql.PostgresConnector",
			"database.hostname": "pg-1",
		})
	}

	t.Run("201 response returns starting connector", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		connector, err := create(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			writeJSON(t, w, map[string]any{
				"name": "pg-connector",
				"config": map[string]string{
					"connector.class":   "io.debezium.connector.postgresql.PostgresConnector",
					"database.hostname": "pg-1",
					"database.dbname":   "testdb",
				},
				"tasks": []any{},
			})
		})

		must.NoError(err)
		must.Equal(
			domain.Connector{
				Name:       "pg-connector",
				PluginType: "io.debezium.connector.postgresql.PostgresConnector",
				Status:     domain.ConnectorStatusStarting,
				TasksCount: 0,
				Tasks:      []domain.Task{},
				Config: domain.SourceConfig{
					Hostname: "pg-1",
					DBName:   strPtr("testdb"),
				},
			},
			connector,
		)
	})

	t.Run("already exists maps to connector already exists", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		_, err := create(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusConflict)
			writeJSON(t, w, map[string]string{
				"error_code": "409",
				"message":    "Connector pg-connector already exists",
			})
		})

		must.ErrorIs(err, domain.ErrConnectorAlreadyExists)
	})

	t.Run("invalid config maps to invalid connector config", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		_, err := create(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(t, w, map[string]any{
				"error_code": 400,
				"message":    "Connector configuration is invalid",
			})
		})

		must.ErrorIs(err, domain.ErrInvalidConnectorConfig)
	})

	t.Run("server error maps to kafka connect unavailable", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		_, err := create(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		must.ErrorIs(err, domain.ErrKafkaConnectUnavailable)
	})

	t.Run("rebalance conflict retries then succeeds", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		calls := 0

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls++

			if calls < 2 {
				w.WriteHeader(http.StatusConflict)
				writeJSON(t, w, map[string]string{
					"error_code": "409",
					"message":    "Request cannot be completed because a rebalance is expected",
				})

				return
			}

			w.WriteHeader(http.StatusCreated)
			writeJSON(t, w, map[string]any{
				"name": "pg-connector",
				"config": map[string]string{
					"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
				},
				"tasks": []any{},
			})
		}))
		t.Cleanup(srv.Close)

		client := newTestClientOnServer(t, srv)

		connector, err := client.CreateConnector(context.Background(), "pg-connector", map[string]string{
			"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
		})

		must.NoError(err)
		must.Equal(2, calls)
		must.Equal(
			domain.Connector{
				Name:       "pg-connector",
				PluginType: "io.debezium.connector.postgresql.PostgresConnector",
				Status:     domain.ConnectorStatusStarting,
				TasksCount: 0,
				Tasks:      []domain.Task{},
				Config:     domain.SourceConfig{},
			},
			connector,
		)
	})
}

func TestHTTPClient_UpdateConnector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler func(w http.ResponseWriter, r *http.Request)
		wantIs  error
	}{
		{
			name: "success puts config and returns fresh connector",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector":
					writeJSON(t, w, map[string]any{
						"name": "pg-connector",
						"config": map[string]string{
							"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
						},
					})
				case r.Method == http.MethodPut && r.URL.Path == "/connectors/pg-connector/config":
					w.WriteHeader(http.StatusOK)
					writeJSON(t, w, map[string]any{
						"name": "pg-connector",
						"config": map[string]string{
							"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
						},
					})
				case r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector/status":
					writeJSON(t, w, map[string]any{
						"connector": map[string]any{"state": "RUNNING", "worker_id": "worker-1"},
						"tasks": []map[string]any{
							{"id": 0, "state": "RUNNING", "worker_id": "worker-1"},
						},
					})
				default:
					http.NotFound(w, nil)
				}
			},
		},
		{
			name: "missing connector on probe maps to connector not found",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				http.NotFound(w, nil)
			},
			wantIs: domain.ErrConnectorNotFound,
		},
		{
			name: "invalid config on put maps to invalid connector config",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector":
					writeJSON(t, w, map[string]any{
						"name": "pg-connector",
						"config": map[string]string{
							"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
						},
					})
				default:
					w.WriteHeader(http.StatusBadRequest)
					writeJSON(t, w, map[string]any{
						"error_code": 400,
						"message":    "Connector configuration is invalid",
					})
				}
			},
			wantIs: domain.ErrInvalidConnectorConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			srv := httptest.NewServer(http.HandlerFunc(tt.handler))
			t.Cleanup(srv.Close)

			client := newTestClientOnServer(t, srv)

			connector, err := client.UpdateConnector(context.Background(), "pg-connector", map[string]string{
				"connector.class":   "io.debezium.connector.postgresql.PostgresConnector",
				"database.hostname": "pg-2",
			})

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			must.Equal(
				domain.Connector{
					Name:       "pg-connector",
					PluginType: "io.debezium.connector.postgresql.PostgresConnector",
					Status:     domain.ConnectorStatusRunning,
					WorkerID:   "worker-1",
					TasksCount: 1,
					Tasks: []domain.Task{
						{ID: 0, State: "running", WorkerID: "worker-1"},
					},
					Config: domain.SourceConfig{},
				},
				connector,
			)
		})
	}
}

func TestHTTPClient_PauseConnector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler func(w http.ResponseWriter, r *http.Request)
		wantIs  error
	}{
		{
			name: "202 response returns connector from status",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodPut && r.URL.Path == "/connectors/pg-connector/pause":
					w.WriteHeader(http.StatusAccepted)
				case r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector":
					writeJSON(t, w, map[string]any{
						"name": "pg-connector",
						"config": map[string]string{
							"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
						},
					})
				case r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector/status":
					writeJSON(t, w, map[string]any{
						"connector": map[string]any{"state": "PAUSED", "worker_id": "worker-1"},
						"tasks":     []map[string]any{},
					})
				default:
					http.NotFound(w, nil)
				}
			},
		},
		{
			name: "missing connector maps to connector not found",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				http.NotFound(w, nil)
			},
			wantIs: domain.ErrConnectorNotFound,
		},
		{
			name: "server error maps to kafka connect unavailable",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantIs: domain.ErrKafkaConnectUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			srv := httptest.NewServer(http.HandlerFunc(tt.handler))
			t.Cleanup(srv.Close)

			client := newTestClientOnServer(t, srv)

			connector, err := client.PauseConnector(context.Background(), "pg-connector")

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			must.Equal(
				domain.Connector{
					Name:       "pg-connector",
					PluginType: "io.debezium.connector.postgresql.PostgresConnector",
					Status:     domain.ConnectorStatusPaused,
					WorkerID:   "worker-1",
					TasksCount: 0,
					Tasks:      []domain.Task{},
					Config:     domain.SourceConfig{},
				},
				connector,
			)
		})
	}
}

func TestHTTPClient_ResumeConnector(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/connectors/pg-connector/resume":
			w.WriteHeader(http.StatusAccepted)
		case r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector":
			writeJSON(t, w, map[string]any{
				"name": "pg-connector",
				"config": map[string]string{
					"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector/status":
			writeJSON(t, w, map[string]any{
				"connector": map[string]any{"state": "RUNNING", "worker_id": "worker-1"},
				"tasks": []map[string]any{
					{"id": 0, "state": "RUNNING", "worker_id": "worker-1"},
				},
			})
		default:
			http.NotFound(w, nil)
		}
	}))
	t.Cleanup(srv.Close)

	client := newTestClientOnServer(t, srv)

	connector, err := client.ResumeConnector(context.Background(), "pg-connector")

	must.NoError(err)
	must.Equal(
		domain.Connector{
			Name:       "pg-connector",
			PluginType: "io.debezium.connector.postgresql.PostgresConnector",
			Status:     domain.ConnectorStatusRunning,
			WorkerID:   "worker-1",
			TasksCount: 1,
			Tasks: []domain.Task{
				{ID: 0, State: "running", WorkerID: "worker-1"},
			},
			Config: domain.SourceConfig{},
		},
		connector,
	)
}

func TestHTTPClient_RestartConnector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler func(w http.ResponseWriter, r *http.Request)
		wantIs  error
	}{
		{
			name: "204 response returns connector from status",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodPost && r.URL.Path == "/connectors/pg-connector/restart":
					w.WriteHeader(http.StatusNoContent)
				case r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector":
					writeJSON(t, w, map[string]any{
						"name": "pg-connector",
						"config": map[string]string{
							"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
						},
					})
				case r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector/status":
					writeJSON(t, w, map[string]any{
						"connector": map[string]any{"state": "RUNNING", "worker_id": "worker-1"},
						"tasks": []map[string]any{
							{"id": 0, "state": "RUNNING", "worker_id": "worker-1"},
						},
					})
				default:
					http.NotFound(w, nil)
				}
			},
		},
		{
			name: "missing connector maps to connector not found",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				http.NotFound(w, nil)
			},
			wantIs: domain.ErrConnectorNotFound,
		},
		{
			name: "server error maps to kafka connect unavailable",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantIs: domain.ErrKafkaConnectUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			srv := httptest.NewServer(http.HandlerFunc(tt.handler))
			t.Cleanup(srv.Close)

			client := newTestClientOnServer(t, srv)

			connector, err := client.RestartConnector(context.Background(), "pg-connector", true, false)

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			must.Equal(
				domain.Connector{
					Name:       "pg-connector",
					PluginType: "io.debezium.connector.postgresql.PostgresConnector",
					Status:     domain.ConnectorStatusRunning,
					WorkerID:   "worker-1",
					TasksCount: 1,
					Tasks: []domain.Task{
						{ID: 0, State: "running", WorkerID: "worker-1"},
					},
					Config: domain.SourceConfig{},
				},
				connector,
			)
		})
	}
}

func TestHTTPClient_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantIs  error
	}{
		{
			name: "204 response deletes connector",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
		},
		{
			name: "missing connector maps to connector not found",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				http.NotFound(w, nil)
			},
			wantIs: domain.ErrConnectorNotFound,
		},
		{
			name: "rebalance exhausts retries and maps to domain error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusConflict)
				writeJSON(t, w, map[string]string{
					"error_code": "409",
					"message":    "Request cannot be completed because a rebalance is expected",
				})
			},
			wantIs: domain.ErrRebalanceInProgress,
		},
		{
			name: "rebalance retries then succeeds",
			handler: func() http.HandlerFunc {
				attempts := 0

				return func(w http.ResponseWriter, _ *http.Request) {
					attempts++
					if attempts < 3 {
						w.WriteHeader(http.StatusConflict)
						writeJSON(t, w, map[string]string{
							"error_code": "409",
							"message":    "Request cannot be completed because a rebalance is expected",
						})

						return
					}

					w.WriteHeader(http.StatusNoContent)
				}
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			client := newTestClient(t, tt.handler)

			err := client.Delete(context.Background(), "pg-connector")

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
		})
	}
}
