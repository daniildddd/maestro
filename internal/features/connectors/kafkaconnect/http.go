package kafkaconnect

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/daniildddd/maestro/internal/core/domain"
)

//nolint:unparam // body is used by create/update requests
func (c *HTTPClient) do(ctx context.Context, method, path string, body, out any) error {
	const op = "connectors.kafkaconnect.do"

	var reader io.Reader

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("%s: marshal body: %w", op, err)
		}

		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("%s: build request: %w", op, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return fmt.Errorf("%s: %w", op, err)
		}

		return fmt.Errorf("%s: %w: %v", op, domain.ErrKafkaConnectUnavailable, err)
	}

	//nolint:errcheck // body fully read below; close error not actionable
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s: read body: %w", op, err)
	}

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		if out != nil && len(data) > 0 {
			if err = json.Unmarshal(data, out); err != nil {
				return fmt.Errorf("%s: unmarshal response: %w", op, err)
			}
		}

		return nil
	case resp.StatusCode >= 500:
		return fmt.Errorf(
			"%s: %w: %w",
			op,
			domain.ErrKafkaConnectUnavailable,
			&HTTPError{StatusCode: resp.StatusCode, Body: string(data)},
		)
	default:
		return fmt.Errorf("%s: %w", op, &HTTPError{StatusCode: resp.StatusCode, Body: string(data)})
	}
}

func (c *HTTPClient) withRetry(
	ctx context.Context,
	retryable func(error) bool,
	fn func() error,
) error {
	var err error

	for attempt := 1; attempt <= c.retryMaxAttempts; attempt++ {
		if err = fn(); err == nil {
			return nil
		}

		if !retryable(err) || attempt == c.retryMaxAttempts {
			break
		}

		delay := c.retryWait(attempt)

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return err
}

func (c *HTTPClient) Ping(ctx context.Context) error {
	const op = "connectors.kafkaconnect.Ping"

	if err := c.do(ctx, http.MethodGet, "/", nil, nil); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *HTTPClient) retryWait(attempt int) time.Duration {
	base := c.retryInitialDelay << (attempt - 1)

	span := base / 5
	if span <= 0 {
		return base
	}

	//nolint:gosec // jitter is not security-sensitive
	jitter := rand.Int63n(int64(2*span)+1) - int64(span)

	return base + time.Duration(jitter)
}
