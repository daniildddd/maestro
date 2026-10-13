package kafkaconnect_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	gotest "github.com/allure-framework/allure-go/commons/gotest" //nolint:revive // package clause is allure, gotest keeps call sites readable
	allurerequire "github.com/allure-framework/allure-go/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func TestHTTPClient_GetConnectors(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "get connectors", func(a *gotest.Context) {
		a.Step("list returns connectors of any engine without failures", func(a *gotest.Context) {
			must := allurerequire.New(a)

			srv := newAllureTestServer(a, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

				writeJSON(a.T(), w, response)
			}))

			client := newTestClientOnServer(a.T(), srv)

			connectors, err := client.GetConnectors(a.Context())

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
							DBName:     new("testdb"),
							PluginName: new("pgoutput"),
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

		a.Step("invalid entry fails with invalid connector", func(a *gotest.Context) {
			must := allurerequire.New(a)

			srv := newAllureTestServer(a, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(a.T(), w, listResponse{
					"broken": newExpansion("", "RUNNING", map[string]string{
						"key.converter": "org.apache.kafka.connect.json.JsonConverter",
					}),
				})
			}))

			client := newTestClientOnServer(a.T(), srv)

			_, err := client.GetConnectors(a.Context())

			must.ErrorIs(err, domain.ErrInvalidConnector)
		})

		a.Step("server error maps to kafka connect unavailable", func(a *gotest.Context) {
			must := allurerequire.New(a)

			srv := newAllureTestServer(a, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}))

			client := newTestClientOnServer(a.T(), srv)

			_, err := client.GetConnectors(a.Context())

			must.ErrorIs(err, domain.ErrKafkaConnectUnavailable)
		})

		a.Step("canceled context returns context canceled", func(a *gotest.Context) {
			must := allurerequire.New(a)

			ctx, cancel := context.WithCancel(a.Context())
			a.T().Cleanup(cancel)

			srv := newAllureTestServer(a, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				cancel()

				select {
				case <-r.Context().Done():
				case <-time.After(2 * time.Second):
					a.T().Error("request context was not canceled")
				}
			}))

			client := newTestClientOnServer(a.T(), srv)

			_, err := client.GetConnectors(ctx)

			must.ErrorIs(err, context.Canceled)
			must.NotErrorIs(err, domain.ErrKafkaConnectUnavailable)
		})
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Get connectors"),
		gotest.WithSeverity("normal"),
		gotest.WithOwner("qa-intern"),
	)
}

func TestHTTPClient_GetConnectorByID(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "get connector by id", func(a *gotest.Context) {
		must := allurerequire.New(a)

		srv := newAllureTestServer(a, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/connectors/pg-connector":
				writeJSON(a.T(), w, map[string]any{
					"name": "pg-connector",
					"config": map[string]string{
						"connector.class":   "io.debezium.connector.postgresql.PostgresConnector",
						"database.hostname": "pg-1",
						"database.dbname":   "testdb",
					},
				})
			case "/connectors/pg-connector/status":
				writeJSON(a.T(), w, map[string]any{
					"connector": map[string]any{"state": "RUNNING", "worker_id": "worker-1"},
					"tasks": []map[string]any{
						{"id": 0, "state": "RUNNING", "worker_id": "worker-1"},
					},
				})
			default:
				http.NotFound(w, nil)
			}
		}))

		client := newTestClientOnServer(a.T(), srv)

		connector, err := client.GetConnectorByID(a.Context(), "pg-connector")

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
					DBName:   new("testdb"),
				},
			},
			connector,
		)
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Get connector"),
		gotest.WithSeverity("normal"),
		gotest.WithOwner("qa-intern"),
	)
}

func TestHTTPClient_CreateConnector(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "create connector", func(a *gotest.Context) {
		create := func(a *gotest.Context, handler func(w http.ResponseWriter, r *http.Request)) (domain.Connector, error) {
			a.T().Helper()

			srv := newAllureTestServer(a, http.HandlerFunc(handler))

			client := newTestClientOnServer(a.T(), srv)

			return client.CreateConnector(a.Context(), "pg-connector", map[string]string{
				"connector.class":   "io.debezium.connector.postgresql.PostgresConnector",
				"database.hostname": "pg-1",
			})
		}

		a.Step("201 response returns starting connector", func(a *gotest.Context) {
			must := allurerequire.New(a)

			connector, err := create(a, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusCreated)
				writeJSON(a.T(), w, map[string]any{
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
						DBName:   new("testdb"),
					},
				},
				connector,
			)
		})

		a.Step("already exists maps to connector already exists", func(a *gotest.Context) {
			must := allurerequire.New(a)

			_, err := create(a, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusConflict)
				writeJSON(a.T(), w, map[string]string{
					"error_code": "409",
					"message":    "Connector pg-connector already exists",
				})
			})

			must.ErrorIs(err, domain.ErrConnectorAlreadyExists)
		})

		a.Step("invalid config maps to invalid connector config", func(a *gotest.Context) {
			must := allurerequire.New(a)

			_, err := create(a, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				writeJSON(a.T(), w, map[string]any{
					"error_code": 400,
					"message":    "Connector configuration is invalid",
				})
			})

			must.ErrorIs(err, domain.ErrInvalidConnectorConfig)
		})

		a.Step("server error maps to kafka connect unavailable", func(a *gotest.Context) {
			must := allurerequire.New(a)

			_, err := create(a, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			})

			must.ErrorIs(err, domain.ErrKafkaConnectUnavailable)
		})

		a.Step("rebalance conflict retries then succeeds", func(a *gotest.Context) {
			must := allurerequire.New(a)

			calls := 0

			srv := newAllureTestServer(a, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls++

				if calls < 2 {
					w.WriteHeader(http.StatusConflict)
					writeJSON(a.T(), w, map[string]string{
						"error_code": "409",
						"message":    "Request cannot be completed because a rebalance is expected",
					})

					return
				}

				w.WriteHeader(http.StatusCreated)
				writeJSON(a.T(), w, map[string]any{
					"name": "pg-connector",
					"config": map[string]string{
						"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
					},
					"tasks": []any{},
				})
			}))

			client := newTestClientOnServer(a.T(), srv)

			connector, err := client.CreateConnector(a.Context(), "pg-connector", map[string]string{
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
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Create connector"),
		gotest.WithSeverity("critical"),
		gotest.WithOwner("qa-intern"),
	)
}

func TestHTTPClient_UpdateConnector(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "update connector", func(a *gotest.Context) {
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
						writeJSON(a.T(), w, map[string]any{
							"name": "pg-connector",
							"config": map[string]string{
								"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
							},
						})
					case r.Method == http.MethodPut && r.URL.Path == "/connectors/pg-connector/config":
						w.WriteHeader(http.StatusOK)
						writeJSON(a.T(), w, map[string]any{
							"name": "pg-connector",
							"config": map[string]string{
								"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
							},
						})
					case r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector/status":
						writeJSON(a.T(), w, map[string]any{
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
						writeJSON(a.T(), w, map[string]any{
							"name": "pg-connector",
							"config": map[string]string{
								"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
							},
						})
					default:
						w.WriteHeader(http.StatusBadRequest)
						writeJSON(a.T(), w, map[string]any{
							"error_code": 400,
							"message":    "Connector configuration is invalid",
						})
					}
				},
				wantIs: domain.ErrInvalidConnectorConfig,
			},
			{
				name:    "rebalance 500 on final status maps to rebalance in progress",
				handler: updateHandler(a.T(), &flakyStatus{failFirst: 1 << 30}),
				wantIs:  domain.ErrRebalanceInProgress,
			},
		}

		for _, tt := range tests {
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

				srv := newAllureTestServer(a, http.HandlerFunc(tt.handler))

				client := newTestClientOnServer(a.T(), srv)

				connector, err := client.UpdateConnector(a.Context(), "pg-connector", map[string]string{
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

		a.Step("rebalance 500 on final status retries then succeeds", func(a *gotest.Context) {
			must := allurerequire.New(a)

			status := &flakyStatus{failFirst: 1}

			srv := newAllureTestServer(a, updateHandler(a.T(), status))

			client := newTestClientOnServer(a.T(), srv)

			connector, err := client.UpdateConnector(a.Context(), "pg-connector", map[string]string{
				"connector.class":   "io.debezium.connector.postgresql.PostgresConnector",
				"database.hostname": "pg-2",
			})

			must.NoError(err)
			must.Equal(2, status.Calls())
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

		a.Step("400 mentioning rebalance maps to invalid config without retry", func(a *gotest.Context) {
			must := allurerequire.New(a)

			var putCalls int

			srv := newAllureTestServer(a, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if serveUpdateProbe(a.T(), w, r) {
					return
				}

				putCalls++

				w.WriteHeader(http.StatusBadRequest)
				writeJSON(a.T(), w, map[string]any{
					"error_code": 400,
					"message":    "Connector configuration is invalid: rebalance is expected after fix",
				})
			}))

			client := newTestClientOnServer(a.T(), srv)

			_, err := client.UpdateConnector(a.Context(), "pg-connector", map[string]string{
				"connector.class":   "io.debezium.connector.postgresql.PostgresConnector",
				"database.hostname": "pg-2",
			})

			must.ErrorIs(err, domain.ErrInvalidConnectorConfig)
			must.Equal(1, putCalls)
		})
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Update connector"),
		gotest.WithSeverity("critical"),
		gotest.WithOwner("qa-intern"),
	)
}

func TestHTTPClient_PauseConnector(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "pause connector", func(a *gotest.Context) {
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
						writeJSON(a.T(), w, map[string]any{
							"name": "pg-connector",
							"config": map[string]string{
								"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
							},
						})
					case r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector/status":
						writeJSON(a.T(), w, map[string]any{
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
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

				srv := newAllureTestServer(a, http.HandlerFunc(tt.handler))

				client := newTestClientOnServer(a.T(), srv)

				connector, err := client.PauseConnector(a.Context(), "pg-connector")

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
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Pause connector"),
		gotest.WithSeverity("critical"),
		gotest.WithOwner("qa-intern"),
	)
}

func TestHTTPClient_ResumeConnector(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "resume connector", func(a *gotest.Context) {
		must := allurerequire.New(a)

		srv := newAllureTestServer(a, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPut && r.URL.Path == "/connectors/pg-connector/resume":
				w.WriteHeader(http.StatusAccepted)
			case r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector":
				writeJSON(a.T(), w, map[string]any{
					"name": "pg-connector",
					"config": map[string]string{
						"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
					},
				})
			case r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector/status":
				writeJSON(a.T(), w, map[string]any{
					"connector": map[string]any{"state": "RUNNING", "worker_id": "worker-1"},
					"tasks": []map[string]any{
						{"id": 0, "state": "RUNNING", "worker_id": "worker-1"},
					},
				})
			default:
				http.NotFound(w, nil)
			}
		}))

		client := newTestClientOnServer(a.T(), srv)

		connector, err := client.ResumeConnector(a.Context(), "pg-connector")

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
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Resume connector"),
		gotest.WithSeverity("critical"),
		gotest.WithOwner("qa-intern"),
	)
}

func TestHTTPClient_RestartConnector(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "restart connector", func(a *gotest.Context) {
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
						writeJSON(a.T(), w, map[string]any{
							"name": "pg-connector",
							"config": map[string]string{
								"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
							},
						})
					case r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector/status":
						writeJSON(a.T(), w, map[string]any{
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
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

				srv := newAllureTestServer(a, http.HandlerFunc(tt.handler))

				client := newTestClientOnServer(a.T(), srv)

				connector, err := client.RestartConnector(a.Context(), "pg-connector", true, false)

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
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Restart connector"),
		gotest.WithSeverity("critical"),
		gotest.WithOwner("qa-intern"),
	)
}

func TestHTTPClient_Delete(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "delete connector", func(a *gotest.Context) {
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
					writeJSON(a.T(), w, map[string]string{
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
							writeJSON(a.T(), w, map[string]string{
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
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

				srv := newAllureTestServer(a, tt.handler)

				client := newTestClientOnServer(a.T(), srv)

				err := client.Delete(a.Context(), "pg-connector")

				if tt.wantIs != nil {
					must.ErrorIs(err, tt.wantIs)

					return
				}

				must.NoError(err)
			})
		}
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Delete connector"),
		gotest.WithSeverity("critical"),
		gotest.WithOwner("qa-intern"),
	)
}

func TestHTTPClient_GetConnectorByID_Errors(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "get connector by id errors", func(a *gotest.Context) {
		tests := []struct {
			name    string
			handler func(w http.ResponseWriter, r *http.Request)
			want    domain.Connector
			wantIs  error
		}{
			{
				name: "success returns full connector",
				handler: func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/connectors/pg-connector":
						writeJSON(a.T(), w, map[string]any{
							"name": "pg-connector",
							"config": map[string]string{
								"connector.class":   "io.debezium.connector.postgresql.PostgresConnector",
								"database.hostname": "pg-1",
								"database.dbname":   "testdb",
							},
						})
					case "/connectors/pg-connector/status":
						writeJSON(a.T(), w, map[string]any{
							"connector": map[string]any{"state": "RUNNING", "worker_id": "worker-1"},
							"tasks": []map[string]any{
								{"id": 0, "state": "RUNNING", "worker_id": "worker-1"},
							},
						})
					default:
						http.NotFound(w, nil)
					}
				},
				want: domain.Connector{
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
						DBName:   new("testdb"),
					},
				},
			},
			{
				name: "missing info maps to not found",
				handler: func(w http.ResponseWriter, _ *http.Request) {
					http.NotFound(w, nil)
				},
				wantIs: domain.ErrConnectorNotFound,
			},
			{
				name: "status error maps to unavailable",
				handler: func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/connectors/pg-connector":
						writeJSON(a.T(), w, map[string]any{
							"name": "pg-connector",
							"config": map[string]string{
								"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
							},
						})
					default:
						w.WriteHeader(http.StatusInternalServerError)
					}
				},
				wantIs: domain.ErrKafkaConnectUnavailable,
			},
			{
				name: "missing plugin maps to invalid",
				handler: func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/connectors/pg-connector":
						writeJSON(a.T(), w, map[string]any{
							"name":   "pg-connector",
							"config": map[string]string{},
						})
					case "/connectors/pg-connector/status":
						writeJSON(a.T(), w, map[string]any{
							"connector": map[string]any{"state": "RUNNING", "worker_id": "worker-1"},
							"tasks":     []map[string]any{},
						})
					default:
						http.NotFound(w, nil)
					}
				},
				wantIs: domain.ErrInvalidConnector,
			},
		}

		for _, tt := range tests {
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

				srv := newAllureTestServer(a, http.HandlerFunc(tt.handler))

				client := newTestClientOnServer(a.T(), srv)

				got, err := client.GetConnectorByID(a.Context(), "pg-connector")

				if tt.wantIs != nil {
					must.ErrorIs(err, tt.wantIs)

					return
				}

				must.NoError(err)
				must.Equal(tt.want, got)
			})
		}
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Get connector"),
		gotest.WithSeverity("normal"),
		gotest.WithOwner("qa-intern"),
	)
}

func TestHTTPClient_CreateConnector_Validation(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "create connector validation", func(a *gotest.Context) {
		tests := []struct {
			name    string
			handler func(w http.ResponseWriter, _ *http.Request)
			want    domain.Connector
			wantIs  error
		}{
			{
				name: "201 returns starting connector",
				handler: func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusCreated)
					writeJSON(a.T(), w, map[string]any{
						"name": "pg-connector",
						"config": map[string]string{
							"connector.class":   "io.debezium.connector.postgresql.PostgresConnector",
							"database.hostname": "pg-1",
							"database.dbname":   "testdb",
						},
					})
				},
				want: domain.Connector{
					Name:       "pg-connector",
					PluginType: "io.debezium.connector.postgresql.PostgresConnector",
					Status:     domain.ConnectorStatusStarting,
					TasksCount: 0,
					Tasks:      []domain.Task{},
					Config: domain.SourceConfig{
						Hostname: "pg-1",
						DBName:   new("testdb"),
					},
				},
			},
			{
				name: "201 without plugin maps to invalid",
				handler: func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusCreated)
					writeJSON(a.T(), w, map[string]any{
						"name":   "pg-connector",
						"config": map[string]string{},
					})
				},
				wantIs: domain.ErrInvalidConnector,
			},
			{
				name: "201 without name maps to invalid",
				handler: func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusCreated)
					writeJSON(a.T(), w, map[string]any{
						"name": "",
						"config": map[string]string{
							"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
						},
					})
				},
				wantIs: domain.ErrInvalidConnector,
			},
		}

		for _, tt := range tests {
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

				srv := newAllureTestServer(a, http.HandlerFunc(tt.handler))

				client := newTestClientOnServer(a.T(), srv)

				got, err := client.CreateConnector(a.Context(), "pg-connector", map[string]string{
					"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
				})

				if tt.wantIs != nil {
					must.ErrorIs(err, tt.wantIs)

					return
				}

				must.NoError(err)
				must.Equal(tt.want, got)
			})
		}
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Create connector"),
		gotest.WithSeverity("critical"),
		gotest.WithOwner("qa-intern"),
	)
}

func serveRestartRunning(t *testing.T) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func TestHTTPClient_RestartConnector_TaskFlags(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "restart connector task flags", func(a *gotest.Context) {
		tests := []struct {
			name         string
			includeTasks bool
			onlyFailed   bool
			want         domain.Connector
		}{
			{
				name:         "204 with tasks returns running",
				includeTasks: true,
				want: domain.Connector{
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
			},
			{
				name:       "204 with only failed returns running",
				onlyFailed: true,
				want: domain.Connector{
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
			},
			{
				name:         "204 with all flags returns running",
				includeTasks: true,
				onlyFailed:   true,
				want: domain.Connector{
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
			},
			{
				name: "204 without flags returns running",
				want: domain.Connector{
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
			},
		}

		for _, tt := range tests {
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

				srv := newAllureTestServer(a, serveRestartRunning(a.T()))

				client := newTestClientOnServer(a.T(), srv)

				got, err := client.RestartConnector(a.Context(), "pg-connector", tt.includeTasks, tt.onlyFailed)

				must.NoError(err)
				must.Equal(tt.want, got)
			})
		}
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Restart connector"),
		gotest.WithSeverity("critical"),
		gotest.WithOwner("qa-intern"),
	)
}

func TestHTTPClient_RestartConnector_RereadErrors(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "restart connector reread errors", func(a *gotest.Context) {
		tests := []struct {
			name    string
			handler func(w http.ResponseWriter, r *http.Request)
			wantIs  error
		}{
			{
				name: "reread missing maps to not found",
				handler: func(w http.ResponseWriter, r *http.Request) {
					if r.Method == http.MethodPost && r.URL.Path == "/connectors/pg-connector/restart" {
						w.WriteHeader(http.StatusNoContent)

						return
					}

					http.NotFound(w, nil)
				},
				wantIs: domain.ErrConnectorNotFound,
			},
			{
				name: "reread invalid maps to invalid",
				handler: func(w http.ResponseWriter, r *http.Request) {
					switch {
					case r.Method == http.MethodPost && r.URL.Path == "/connectors/pg-connector/restart":
						w.WriteHeader(http.StatusNoContent)
					case r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector":
						writeJSON(a.T(), w, map[string]any{
							"name":   "pg-connector",
							"config": map[string]string{},
						})
					case r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector/status":
						writeJSON(a.T(), w, map[string]any{
							"connector": map[string]any{"state": "RUNNING", "worker_id": "worker-1"},
							"tasks":     []map[string]any{},
						})
					default:
						http.NotFound(w, nil)
					}
				},
				wantIs: domain.ErrInvalidConnector,
			},
		}

		for _, tt := range tests {
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

				srv := newAllureTestServer(a, http.HandlerFunc(tt.handler))

				client := newTestClientOnServer(a.T(), srv)

				_, err := client.RestartConnector(a.Context(), "pg-connector", false, false)

				must.ErrorIs(err, tt.wantIs)
			})
		}
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Restart connector"),
		gotest.WithSeverity("critical"),
		gotest.WithOwner("qa-intern"),
	)
}

func servePauseOK(t *testing.T) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func TestHTTPClient_PauseConnector_Reread(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "pause connector reread", func(a *gotest.Context) {
		tests := []struct {
			name string
			want domain.Connector
		}{
			{
				name: "202 returns paused connector",
				want: domain.Connector{
					Name:       "pg-connector",
					PluginType: "io.debezium.connector.postgresql.PostgresConnector",
					Status:     domain.ConnectorStatusPaused,
					WorkerID:   "worker-1",
					TasksCount: 0,
					Tasks:      []domain.Task{},
					Config:     domain.SourceConfig{},
				},
			},
		}

		for _, tt := range tests {
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

				srv := newAllureTestServer(a, servePauseOK(a.T()))

				client := newTestClientOnServer(a.T(), srv)

				got, err := client.PauseConnector(a.Context(), "pg-connector")

				must.NoError(err)
				must.Equal(tt.want, got)
			})
		}
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Pause connector"),
		gotest.WithSeverity("critical"),
		gotest.WithOwner("qa-intern"),
	)
}

func TestHTTPClient_PauseConnector_RereadErrors(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "pause connector reread errors", func(a *gotest.Context) {
		tests := []struct {
			name    string
			handler func(w http.ResponseWriter, r *http.Request)
			wantIs  error
		}{
			{
				name: "reread missing maps to not found",
				handler: func(w http.ResponseWriter, r *http.Request) {
					if r.Method == http.MethodPut && r.URL.Path == "/connectors/pg-connector/pause" {
						w.WriteHeader(http.StatusAccepted)

						return
					}

					http.NotFound(w, nil)
				},
				wantIs: domain.ErrConnectorNotFound,
			},
			{
				name: "reread invalid maps to invalid",
				handler: func(w http.ResponseWriter, r *http.Request) {
					switch {
					case r.Method == http.MethodPut && r.URL.Path == "/connectors/pg-connector/pause":
						w.WriteHeader(http.StatusAccepted)
					case r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector":
						writeJSON(a.T(), w, map[string]any{
							"name":   "pg-connector",
							"config": map[string]string{},
						})
					case r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector/status":
						writeJSON(a.T(), w, map[string]any{
							"connector": map[string]any{"state": "PAUSED", "worker_id": "worker-1"},
							"tasks":     []map[string]any{},
						})
					default:
						http.NotFound(w, nil)
					}
				},
				wantIs: domain.ErrInvalidConnector,
			},
		}

		for _, tt := range tests {
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

				srv := newAllureTestServer(a, http.HandlerFunc(tt.handler))

				client := newTestClientOnServer(a.T(), srv)

				_, err := client.PauseConnector(a.Context(), "pg-connector")

				must.ErrorIs(err, tt.wantIs)
			})
		}
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Pause connector"),
		gotest.WithSeverity("critical"),
		gotest.WithOwner("qa-intern"),
	)
}
