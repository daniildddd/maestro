package collector_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/features/connectors/collector"
)

type stubLister struct {
	mu sync.Mutex

	connectors []domain.Connector
	err        error
	calls      int

	waitForContext bool
}

func (s *stubLister) GetConnectors(ctx context.Context) ([]domain.Connector, error) {
	s.mu.Lock()
	s.calls++
	s.mu.Unlock()

	if s.waitForContext {
		<-ctx.Done()

		return nil, ctx.Err()
	}

	return s.connectors, s.err
}

func (s *stubLister) Calls() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.calls
}

type stubMetrics struct {
	mu sync.Mutex

	connectors []domain.Connector
	calls      int
}

func (s *stubMetrics) SetConnectors(connectors []domain.Connector) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls++
	s.connectors = connectors
}

func (s *stubMetrics) Calls() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.calls
}

func (s *stubMetrics) Connectors() []domain.Connector {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.connectors
}

func newTestCollector(
	lister collector.ConnectorLister,
	metrics collector.ConnectorMetrics,
	cfg collector.CollectorConfig,
) *collector.Collector {
	return collector.NewCollector(
		lister,
		metrics,
		&core_logger.Logger{
			Logger: zap.NewNop(),
		},
		cfg,
	)
}

func TestNewCollector(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	c := newTestCollector(
		&stubLister{},
		&stubMetrics{},
		collector.CollectorConfig{},
	)
	must.NotNil(c)
}

func TestCollector_Run(t *testing.T) {
	t.Parallel()

	t.Run("collects periodically until context cancelled", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		lister := &stubLister{
			connectors: []domain.Connector{
				{
					Name:   "pg-connector",
					Status: domain.ConnectorStatusRunning,
				},
			},
		}

		metrics := &stubMetrics{}

		c := newTestCollector(
			lister,
			metrics,
			collector.CollectorConfig{
				Interval: 10 * time.Millisecond,
				Timeout:  time.Second,
			},
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		done := make(chan struct{})

		go func() {
			defer close(done)

			c.Run(ctx)
		}()

		must.Eventually(func() bool {
			return lister.Calls() >= 2 && metrics.Calls() >= 2
		}, time.Second, 5*time.Millisecond)

		cancel()

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("collector did not stop after context cancel")
		}
	})

	t.Run("stops on context cancel without collecting", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		lister := &stubLister{}
		metrics := &stubMetrics{}

		c := newTestCollector(
			lister,
			metrics,
			collector.CollectorConfig{
				Interval: time.Hour,
				Timeout:  time.Second,
			},
		)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		done := make(chan struct{})

		go func() {
			defer close(done)

			c.Run(ctx)
		}()

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("collector did not stop after context cancel")
		}

		must.Equal(0, lister.Calls())
	})
}

func TestCollector_Collect(t *testing.T) {
	t.Parallel()

	t.Run("success calls SetConnectors with listed connectors", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		connectors := []domain.Connector{
			{
				Name:   "pg-connector",
				Status: domain.ConnectorStatusRunning,
			},
			{
				Name:   "mysql-connector",
				Status: domain.ConnectorStatusPaused,
			},
		}

		lister := &stubLister{
			connectors: connectors,
		}

		metrics := &stubMetrics{}

		c := newTestCollector(
			lister,
			metrics,
			collector.CollectorConfig{
				Timeout: time.Second,
			},
		)

		c.Collect(context.Background())

		must.Equal(1, lister.Calls())
		must.Equal(1, metrics.Calls())
		is.Equal(connectors, metrics.Connectors())
	})

	t.Run("GetConnectors error is logged and SetConnectors skipped", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		lister := &stubLister{
			err: errors.New("get connectors failed"),
		}

		metrics := &stubMetrics{}

		core, logs := observer.New(zapcore.DebugLevel)

		c := collector.NewCollector(
			lister,
			metrics,
			&core_logger.Logger{
				Logger: zap.New(core),
			},
			collector.CollectorConfig{
				Timeout: time.Second,
			},
		)

		c.Collect(context.Background())

		must.Equal(1, lister.Calls())
		must.Equal(0, metrics.Calls())
		is.Empty(metrics.Connectors())

		entries := logs.FilterMessage("collect connector metrics").All()
		must.Len(entries, 1)
		is.Equal("get connectors failed", entries[0].ContextMap()["error"])
	})

	t.Run("hanging lister is cancelled by timeout", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		lister := &stubLister{
			waitForContext: true,
		}

		metrics := &stubMetrics{}

		core, logs := observer.New(zapcore.DebugLevel)

		c := collector.NewCollector(
			lister,
			metrics,
			&core_logger.Logger{
				Logger: zap.New(core),
			},
			collector.CollectorConfig{
				Timeout: 20 * time.Millisecond,
			},
		)

		c.Collect(context.Background())

		must.Equal(1, lister.Calls())
		must.Equal(0, metrics.Calls())

		entries := logs.FilterMessage("collect connector metrics").All()
		must.Len(entries, 1)
		is.Equal("context deadline exceeded", entries[0].ContextMap()["error"])
	})
}
