package kafkaconnect_test

import (
	"context"
	"net/http"
	"testing"

	gotest "github.com/allure-framework/allure-go/commons/gotest" //nolint:revive // package clause is allure, gotest keeps call sites readable
	allurerequire "github.com/allure-framework/allure-go/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func TestHTTPClient_GetTaskByID(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "get task by id", func(a *gotest.Context) {
		tests := []struct {
			name       string
			taskStatus func(w http.ResponseWriter, _ *http.Request)
			want       domain.Task
			wantIs     error
		}{
			{
				name: "success returns task with lowercased state",
				taskStatus: func(w http.ResponseWriter, _ *http.Request) {
					writeJSON(a.T(), w, map[string]any{
						"id": 0, "state": "RUNNING", "worker_id": "worker-1",
						"trace": "worker died: oom",
					})
				},
				want: domain.Task{
					ID:       0,
					State:    "running",
					WorkerID: "worker-1",
					Trace:    new("worker died: oom"),
				},
			},
			{
				name: "missing task maps to connector task not found",
				taskStatus: func(w http.ResponseWriter, _ *http.Request) {
					http.NotFound(w, nil)
				},
				wantIs: domain.ErrConnectorTaskNotFound,
			},
			{
				name: "server error maps to kafka connect unavailable",
				taskStatus: func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				},
				wantIs: domain.ErrKafkaConnectUnavailable,
			},
		}

		for _, tt := range tests {
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

				srv := newAllureTestServer(a, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/connectors/pg-connector/tasks/0/status" {
						tt.taskStatus(w, r)

						return
					}

					http.NotFound(w, nil)
				}))

				client := newTestClientOnServer(a.T(), srv)

				task, err := client.GetTaskByID(a.Context(), "pg-connector", 0)

				if tt.wantIs != nil {
					must.ErrorIs(err, tt.wantIs)

					return
				}

				must.NoError(err)
				must.Equal(tt.want, task)
			})
		}
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Get task"),
		gotest.WithSeverity("normal"),
		gotest.WithOwner("qa-intern"),
	)
}

func TestHTTPClient_RestartTask(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "restart task", func(a *gotest.Context) {
		conflict := func(message string) http.HandlerFunc {
			return func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusConflict)
				writeJSON(a.T(), w, map[string]string{
					"error_code": "409",
					"message":    message,
				})
			}
		}

		a.Step("204 restarts task", func(a *gotest.Context) {
			must := allurerequire.New(a)

			attempts, err := runRestartTask(a.Context(), a, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})

			must.NoError(err)
			must.Equal(1, attempts)
		})

		a.Step("missing task maps to connector task not found", func(a *gotest.Context) {
			must := allurerequire.New(a)

			attempts, err := runRestartTask(a.Context(), a, func(w http.ResponseWriter, _ *http.Request) {
				http.NotFound(w, nil)
			})

			must.ErrorIs(err, domain.ErrConnectorTaskNotFound)
			must.Equal(1, attempts)
		})

		a.Step("server error maps to kafka connect unavailable", func(a *gotest.Context) {
			must := allurerequire.New(a)

			attempts, err := runRestartTask(a.Context(), a, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			})

			must.ErrorIs(err, domain.ErrKafkaConnectUnavailable)
			must.Equal(1, attempts)
		})

		a.Step("rebalance conflict retries then succeeds", func(a *gotest.Context) {
			must := allurerequire.New(a)

			calls := 0

			handler := func(w http.ResponseWriter, r *http.Request) {
				calls++

				if calls < 2 {
					conflict("Request cannot be completed because a rebalance is expected")(w, r)

					return
				}

				w.WriteHeader(http.StatusNoContent)
			}

			attempts, err := runRestartTask(a.Context(), a, handler)

			must.NoError(err)
			must.Equal(2, attempts)
		})

		a.Step("conflict without rebalance word retries", func(a *gotest.Context) {
			must := allurerequire.New(a)

			calls := 0

			handler := func(w http.ResponseWriter, r *http.Request) {
				calls++

				if calls < 2 {
					conflict("Cannot complete request because the worker can't be found")(w, r)

					return
				}

				w.WriteHeader(http.StatusNoContent)
			}

			attempts, err := runRestartTask(a.Context(), a, handler)

			must.NoError(err)
			must.Equal(2, attempts)
		})

		a.Step("rebalance word without conflict status does not retry", func(a *gotest.Context) {
			must := allurerequire.New(a)

			attempts, err := runRestartTask(a.Context(), a, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				writeJSON(a.T(), w, map[string]string{
					"error_code": "400",
					"message":    "Request mentions rebalance but is a bad request",
				})
			})

			must.Error(err)
			must.Equal(1, attempts)
		})

		a.Step("context cancel aborts retry wait", func(a *gotest.Context) {
			must := allurerequire.New(a)

			ctx, cancel := context.WithCancel(a.Context())
			a.T().Cleanup(cancel)

			srv := newAllureTestServer(a, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				cancel()
				conflict("Request cannot be completed because a rebalance is expected")(w, r)
			}))

			client := newTestClientOnServer(a.T(), srv)

			err := client.RestartTask(ctx, "pg-connector", 0)

			must.ErrorIs(err, context.Canceled)
		})
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("Restart task"),
		gotest.WithSeverity("critical"),
		gotest.WithOwner("qa-intern"),
	)
}
