package kafkaconnect

import (
	"net/http"
	"time"

	"github.com/daniildddd/maestro/internal/features/connectors/plugins"
)

type HTTPClient struct {
	baseURL           string
	client            *http.Client
	registry          *plugins.Registry
	retryMaxAttempts  int
	retryInitialDelay time.Duration
}

func NewHTTPClient(cfg Config, registry *plugins.Registry) *HTTPClient {
	return &HTTPClient{
		baseURL:           cfg.BaseURL,
		client:            &http.Client{Timeout: cfg.Timeout},
		registry:          registry,
		retryMaxAttempts:  cfg.RetryMaxAttempts,
		retryInitialDelay: cfg.RetryInitialDelay,
	}
}
