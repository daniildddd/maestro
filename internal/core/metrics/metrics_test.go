package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/metrics"
)

func TestMetrics_Observe(t *testing.T) {
	t.Parallel()

	m := metrics.New(metrics.Config{Buckets: []float64{0.1, 1}})

	m.Observe("GET", "GET /api/v1/users/{id}", 200, 50*time.Millisecond)
	m.Observe("GET", "GET /api/v1/users/{id}", 500, 150*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	body := rec.Body.String()

	assert.Contains(t, body, `http_server_requests_total{method="GET",route="GET /api/v1/users/{id}",status="200"} 1`)
	assert.Contains(t, body, `http_server_requests_total{method="GET",route="GET /api/v1/users/{id}",status="500"} 1`)
	assert.Contains(t, body, "http_server_request_duration_seconds_bucket")
}

func TestMetrics_SetConnectors(t *testing.T) {
	t.Parallel()

	m := metrics.New(metrics.Config{Buckets: []float64{0.1, 1}})

	m.SetConnectors([]domain.Connector{
		{
			Name:   "pg-connector",
			Status: domain.ConnectorStatusFailed,
			Tasks: []domain.Task{
				{ID: 0, State: "FAILED"},
				{ID: 1, State: "RUNNING"},
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, req)

	body := rec.Body.String()

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, body, `maestro_connector_info{connector="pg-connector",state="failed"} 1`)
	assert.Contains(t, body, `maestro_connector_tasks{connector="pg-connector",state="FAILED"} 1`)
	assert.Contains(t, body, `maestro_connector_tasks{connector="pg-connector",state="RUNNING"} 1`)

	m.SetConnectors(nil)

	rec = httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, req)

	assert.NotContains(t, rec.Body.String(), `maestro_connector_info{connector="pg-connector"`)
}
