package kafkaconnect_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func TestHTTPClient_GetTaskByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		taskStatus func(w http.ResponseWriter, _ *http.Request)
		want       domain.Task
		wantIs     error
	}{
		{
			name: "success returns task with lowercased state",
			taskStatus: func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(t, w, map[string]any{
					"id": 0, "state": "RUNNING", "worker_id": "worker-1",
					"trace": "worker died: oom",
				})
			},
			want: domain.Task{
				ID:       0,
				State:    "running",
				WorkerID: "worker-1",
				Trace:    strPtr("worker died: oom"),
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/connectors/pg-connector/tasks/0/status" {
					tt.taskStatus(w, r)

					return
				}

				http.NotFound(w, nil)
			}))
			t.Cleanup(srv.Close)

			client := newTestClientOnServer(t, srv)

			task, err := client.GetTaskByID(context.Background(), "pg-connector", 0)

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			must.Equal(tt.want, task)
		})
	}
}

func TestHTTPClient_RestartTask(t *testing.T) {
	t.Parallel()

	conflict := func(message string) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusConflict)
			writeJSON(t, w, map[string]string{
				"error_code": "409",
				"message":    message,
			})
		}
	}

	t.Run("204 restarts task", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		attempts, err := runRestartTask(context.Background(), t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		must.NoError(err)
		must.Equal(1, attempts)
	})

	t.Run("missing task maps to connector task not found", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		attempts, err := runRestartTask(context.Background(), t, func(w http.ResponseWriter, _ *http.Request) {
			http.NotFound(w, nil)
		})

		must.ErrorIs(err, domain.ErrConnectorTaskNotFound)
		must.Equal(1, attempts)
	})

	t.Run("server error maps to kafka connect unavailable", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		attempts, err := runRestartTask(context.Background(), t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		must.ErrorIs(err, domain.ErrKafkaConnectUnavailable)
		must.Equal(1, attempts)
	})

	t.Run("rebalance conflict retries then succeeds", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		calls := 0

		attempts, err := runRestartTask(context.Background(), t, func(w http.ResponseWriter, r *http.Request) {
			calls++

			if calls < 2 {
				conflict("Request cannot be completed because a rebalance is expected")(w, r)

				return
			}

			w.WriteHeader(http.StatusNoContent)
		})

		must.NoError(err)
		must.Equal(2, attempts)
	})

	t.Run("conflict without rebalance word retries", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		calls := 0

		attempts, err := runRestartTask(context.Background(), t, func(w http.ResponseWriter, r *http.Request) {
			calls++

			if calls < 2 {
				conflict("Cannot complete request because the worker can't be found")(w, r)

				return
			}

			w.WriteHeader(http.StatusNoContent)
		})

		must.NoError(err)
		must.Equal(2, attempts)
	})

	t.Run("rebalance word without conflict status does not retry", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		attempts, err := runRestartTask(context.Background(), t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(t, w, map[string]string{
				"error_code": "400",
				"message":    "Request mentions rebalance but is a bad request",
			})
		})

		must.Error(err)
		must.Equal(1, attempts)
	})

	t.Run("context cancel aborts retry wait", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cancel()
			conflict("Request cannot be completed because a rebalance is expected")(w, r)
		}))
		t.Cleanup(srv.Close)

		client := newTestClientOnServer(t, srv)

		err := client.RestartTask(ctx, "pg-connector", 0)

		must.ErrorIs(err, context.Canceled)
	})
}
