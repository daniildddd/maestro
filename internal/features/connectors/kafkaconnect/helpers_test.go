package kafkaconnect_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/daniildddd/maestro/internal/features/connectors/kafkaconnect"
	"github.com/daniildddd/maestro/internal/features/connectors/plugins"
)

const (
	defaultTestTimeout = time.Second
	retryDelayUnit     = time.Millisecond
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *kafkaconnect.HTTPClient {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	return kafkaconnect.NewHTTPClient(
		kafkaconnect.Config{
			BaseURL:           srv.URL,
			Timeout:           time.Second,
			RetryMaxAttempts:  3,
			RetryInitialDelay: 5 * time.Millisecond,
		},
		plugins.NewRegistry(
			plugins.PostgresAdapter{},
			plugins.MySQLAdapter{},
			plugins.MongoDBAdapter{},
			plugins.SQLServerAdapter{},
			plugins.OracleAdapter{},
		),
	)
}

func newTestClientOnServer(t *testing.T, srv *httptest.Server) *kafkaconnect.HTTPClient {
	t.Helper()

	return kafkaconnect.NewHTTPClient(
		kafkaconnect.Config{
			BaseURL:           srv.URL,
			Timeout:           defaultTestTimeout,
			RetryMaxAttempts:  3,
			RetryInitialDelay: 5 * retryDelayUnit,
		},
		plugins.NewRegistry(
			plugins.PostgresAdapter{},
			plugins.MySQLAdapter{},
			plugins.MongoDBAdapter{},
			plugins.SQLServerAdapter{},
			plugins.OracleAdapter{},
		),
	)
}

func serveUpdateProbe(t *testing.T, w http.ResponseWriter, r *http.Request) bool {
	t.Helper()

	if r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector" {
		writeJSON(t, w, map[string]any{
			"name": "pg-connector",
			"config": map[string]string{
				"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
			},
		})

		return true
	}

	return false
}

type flakyStatus struct {
	failFirst int
	calls     int
}

func (f *flakyStatus) Calls() int {
	return f.calls
}

func (f *flakyStatus) serve(t *testing.T, w http.ResponseWriter) {
	t.Helper()

	f.calls++

	if f.calls <= f.failFirst {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(t, w, map[string]any{
			"error_code": 500,
			"message":    "Request cannot be completed because a rebalance is expected",
		})

		return
	}

	writeJSON(t, w, map[string]any{
		"connector": map[string]any{"state": "RUNNING", "worker_id": "worker-1"},
		"tasks": []map[string]any{
			{"id": 0, "state": "RUNNING", "worker_id": "worker-1"},
		},
	})
}

func updateHandler(t *testing.T, status *flakyStatus) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		if serveUpdateProbe(t, w, r) {
			return
		}

		if r.Method == http.MethodPut && r.URL.Path == "/connectors/pg-connector/config" {
			w.WriteHeader(http.StatusOK)
			writeJSON(t, w, map[string]any{
				"name": "pg-connector",
				"config": map[string]string{
					"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
				},
			})

			return
		}

		if r.Method == http.MethodGet && r.URL.Path == "/connectors/pg-connector/status" {
			status.serve(t, w)

			return
		}

		http.NotFound(w, nil)
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, payload any) {
	t.Helper()

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Errorf("encode response: %v", err)
	}
}

func strPtr(s string) *string { return &s }

type listResponse map[string]expansion

type expansion struct {
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

func newExpansion(class, state string, config map[string]string) expansion {
	var e expansion

	e.Info.Config = config
	e.Info.Config["connector.class"] = class
	e.Status.Connector.State = state
	e.Status.Tasks = []struct {
		ID       int    `json:"id"`
		State    string `json:"state"`
		WorkerID string `json:"worker_id"`
	}{
		{ID: 0, State: state, WorkerID: "worker-1"},
	}

	return e
}

func runRestartTask(ctx context.Context, t *testing.T, handler http.HandlerFunc) (int, error) {
	t.Helper()

	attempts := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++

		handler(w, r)
	}))
	t.Cleanup(srv.Close)

	client := newTestClientOnServer(t, srv)

	return attempts, client.RestartTask(ctx, "pg-connector", 0)
}
