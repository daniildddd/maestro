package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/kelseyhightower/envconfig"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/logger"
)

type CollectorConfig struct {
	Interval time.Duration `default:"15s" envconfig:"INTERVAL"`
	Timeout  time.Duration `default:"10s" envconfig:"TIMEOUT"`
}

func NewCollectorConfigMust() CollectorConfig {
	var cfg CollectorConfig

	if err := envconfig.Process("METRICS_COLLECTOR", &cfg); err != nil {
		panic(fmt.Errorf("process metrics collector config: %w", err))
	}

	return cfg
}

type ConnectorLister interface {
	GetConnectors(ctx context.Context) ([]domain.Connector, error)
}

type ConnectorMetrics interface {
	SetConnectors([]domain.Connector)
}

type Collector struct {
	lister  ConnectorLister
	metrics ConnectorMetrics
	logger  *logger.Logger
	cfg     CollectorConfig
}

func NewCollector(
	lister ConnectorLister,
	m ConnectorMetrics,
	log *logger.Logger,
	cfg CollectorConfig,
) *Collector {
	return &Collector{
		lister:  lister,
		metrics: m,
		logger:  log,
		cfg:     cfg,
	}
}

func (c *Collector) Run(ctx context.Context) {
	ticker := time.NewTicker(c.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.Collect(ctx)
		}
	}
}

func (c *Collector) Collect(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	connectors, err := c.lister.GetConnectors(ctx)
	if err != nil {
		c.logger.Error("collect connector metrics", zap.Error(err))

		return
	}

	c.metrics.SetConnectors(connectors)
}
