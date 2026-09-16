package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/daniildddd/maestro/internal/core/domain"
)

type Metrics struct {
	registry *prometheus.Registry

	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec

	connectorInfo  *prometheus.GaugeVec
	connectorTasks *prometheus.GaugeVec
}

func New(cfg Config) *Metrics {
	reg := prometheus.NewRegistry()

	m := &Metrics{
		registry: reg,
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_server_requests_total",
				Help: "Total HTTP requests by method, route and status.",
			},
			[]string{"method", "route", "status"},
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_server_request_duration_seconds",
				Help:    "HTTP request duration in seconds.",
				Buckets: cfg.Buckets,
			},
			[]string{"method", "route"},
		),
		connectorInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "maestro_connector_info",
				Help: "Current connector state (1 per connector/state).",
			},
			[]string{"connector", "state"},
		),
		connectorTasks: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "maestro_connector_tasks",
				Help: "Number of tasks by connector and state.",
			},
			[]string{"connector", "state"},
		),
	}

	reg.MustRegister(m.requests, m.duration, m.connectorInfo, m.connectorTasks)

	return m
}

func (m *Metrics) Observe(method, route string, status int, d time.Duration) {
	s := strconv.Itoa(status)
	m.requests.WithLabelValues(method, route, s).Inc()
	m.duration.WithLabelValues(method, route).Observe(d.Seconds())
}

func (m *Metrics) SetConnectors(connectors []domain.Connector) {
	m.connectorInfo.Reset()
	m.connectorTasks.Reset()

	for _, c := range connectors {
		m.connectorInfo.WithLabelValues(c.Name, c.Status).Set(1)

		byState := make(map[string]float64, len(c.Tasks))
		for _, t := range c.Tasks {
			byState[t.State]++
		}

		for state, n := range byState {
			m.connectorTasks.WithLabelValues(c.Name, state).Set(n)
		}
	}
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}
