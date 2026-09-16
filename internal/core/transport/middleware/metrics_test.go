package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	core_metrics "github.com/daniildddd/maestro/internal/core/metrics"
	"github.com/daniildddd/maestro/internal/core/transport/middleware"
)

func TestMetrics(t *testing.T) {
	t.Parallel()

	t.Run("observes templated route and status", func(t *testing.T) {
		t.Parallel()

		m := core_metrics.New(core_metrics.Config{Buckets: []float64{0.1, 1}})

		final := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/connectors/pg-connector", http.NoBody)
		req.Pattern = "GET /api/v1/connectors/{id}"
		rec := httptest.NewRecorder()

		middleware.Metrics(m)(final).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusTeapot, rec.Code)
		assertExpositionContains(
			t,
			m,
			`http_server_requests_total{method="GET",route="GET /api/v1/connectors/{id}",status="418"} 1`,
		)
	})

	t.Run("unmatched pattern recorded as unmatched", func(t *testing.T) {
		t.Parallel()

		m := core_metrics.New(core_metrics.Config{Buckets: []float64{0.1, 1}})

		final := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		req := httptest.NewRequest(http.MethodGet, "/nope", http.NoBody)
		rec := httptest.NewRecorder()

		middleware.Metrics(m)(final).ServeHTTP(rec, req)

		assertExpositionContains(
			t,
			m,
			`http_server_requests_total{method="GET",route="unmatched",status="404"} 1`,
		)
	})

	t.Run("metrics endpoint is not observed", func(t *testing.T) {
		t.Parallel()

		m := core_metrics.New(core_metrics.Config{Buckets: []float64{0.1, 1}})

		final := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
		rec := httptest.NewRecorder()

		middleware.Metrics(m)(final).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assertExpositionNotContains(t, m, "http_server_requests_total")
	})
}

func assertExpositionContains(t *testing.T, m *core_metrics.Metrics, want string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), want)
}

func assertExpositionNotContains(t *testing.T, m *core_metrics.Metrics, want string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.NotContains(t, rec.Body.String(), want)
}
