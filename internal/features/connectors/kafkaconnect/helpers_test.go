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
