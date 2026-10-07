package kafkaconnect_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gotest "github.com/allure-framework/allure-go/commons/gotest" //nolint:revive // package clause is allure, gotest keeps call sites readable
	"github.com/allure-framework/allure-go/commons/httpexchange"

	"github.com/daniildddd/maestro/internal/features/connectors/kafkaconnect"
	"github.com/daniildddd/maestro/internal/features/connectors/plugins"
)

const (
	defaultTestTimeout = time.Second
	retryDelayUnit     = time.Millisecond
)

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

func newAllureTestServer(a *gotest.Context, handler http.Handler) *httptest.Server {
	a.T().Helper()

	srv := httptest.NewServer(httpexchange.NewHandler(a.Context(), handler))
	a.T().Cleanup(srv.Close)

	return srv
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

//nolint:contextcheck // context comes from the caller step; creating it here would detach cancellation
func runRestartTask(ctx context.Context, a *gotest.Context, handler http.HandlerFunc) (int, error) {
	a.T().Helper()

	attempts := 0

	srv := newAllureTestServer(a, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++

		handler(w, r)
	}))

	client := newTestClientOnServer(a.T(), srv)

	return attempts, client.RestartTask(ctx, "pg-connector", 0)
}
