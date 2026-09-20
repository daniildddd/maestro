package kafkaconnect

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/features/connectors/plugins"
)

func TestNewHTTPClient(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	registry := plugins.NewRegistry(plugins.PostgresAdapter{})

	client := NewHTTPClient(Config{
		BaseURL:           "http://connect:8083",
		Timeout:           time.Second,
		RetryMaxAttempts:  5,
		RetryInitialDelay: 300 * time.Millisecond,
	}, registry)

	must.Equal("http://connect:8083", client.baseURL)
	must.Equal(time.Second, client.client.Timeout)
	must.Same(registry, client.registry)
	must.Equal(5, client.retryMaxAttempts)
	must.Equal(300*time.Millisecond, client.retryInitialDelay)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type errReader struct {
	err error
}

func (r errReader) Read([]byte) (int, error) {
	return 0, r.err
}

func (r errReader) Close() error {
	return nil
}

func TestHTTPClient_Ping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		status       int
		wantIs       error
		wantContains string
	}{
		{
			name:   "200 returns nil",
			status: http.StatusOK,
		},
		{
			name:   "500 maps to unavailable",
			status: http.StatusInternalServerError,
			wantIs: domain.ErrKafkaConnectUnavailable,
		},
		{
			name:         "404 returns unexpected status",
			status:       http.StatusNotFound,
			wantContains: "unexpected status 404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			var gotPath string

			var gotMethod string

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotMethod = r.Method

				w.WriteHeader(tt.status)
			}))
			t.Cleanup(srv.Close)

			client := NewHTTPClient(Config{
				BaseURL:           srv.URL,
				Timeout:           time.Second,
				RetryMaxAttempts:  3,
				RetryInitialDelay: time.Millisecond,
			}, plugins.NewRegistry(plugins.PostgresAdapter{}))

			err := client.Ping(context.Background())

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			if tt.wantContains != "" {
				must.ErrorContains(err, tt.wantContains)

				return
			}

			must.NoError(err)
			must.Equal("/", gotPath)
			must.Equal(http.MethodGet, gotMethod)
		})
	}
}

func TestHTTPClient_Do(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "200 unmarshals response",
			run: func(t *testing.T) {
				t.Helper()

				must := require.New(t)

				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"name":"pg-connector"}`)) //nolint:errcheck // test response write
				}))
				t.Cleanup(srv.Close)

				client := NewHTTPClient(Config{
					BaseURL:           srv.URL,
					Timeout:           time.Second,
					RetryMaxAttempts:  3,
					RetryInitialDelay: time.Millisecond,
				}, plugins.NewRegistry(plugins.PostgresAdapter{}))

				var got struct {
					Name string `json:"name"`
				}

				err := client.do(context.Background(), http.MethodGet, "/x", nil, &got)

				must.NoError(err)
				must.Equal(struct {
					Name string `json:"name"`
				}{Name: "pg-connector"}, got)
			},
		},
		{
			name: "marshal error",
			run: func(t *testing.T) {
				t.Helper()

				must := require.New(t)

				client := NewHTTPClient(Config{
					BaseURL:           "http://example.com",
					Timeout:           time.Second,
					RetryMaxAttempts:  3,
					RetryInitialDelay: time.Millisecond,
				}, plugins.NewRegistry(plugins.PostgresAdapter{}))

				err := client.do(context.Background(), http.MethodPost, "/x", func() {}, nil)

				must.ErrorContains(err, "marshal body")
			},
		},
		{
			name: "build request error",
			run: func(t *testing.T) {
				t.Helper()

				must := require.New(t)

				client := NewHTTPClient(Config{
					BaseURL:           "://bad-url",
					Timeout:           time.Second,
					RetryMaxAttempts:  3,
					RetryInitialDelay: time.Millisecond,
				}, plugins.NewRegistry(plugins.PostgresAdapter{}))

				err := client.do(context.Background(), http.MethodGet, "/x", nil, nil)

				must.ErrorContains(err, "build request")
			},
		},
		{
			name: "transport error maps to unavailable",
			run: func(t *testing.T) {
				t.Helper()

				must := require.New(t)

				client := NewHTTPClient(Config{
					BaseURL:           "http://example.com",
					Timeout:           time.Second,
					RetryMaxAttempts:  3,
					RetryInitialDelay: time.Millisecond,
				}, plugins.NewRegistry(plugins.PostgresAdapter{}))
				client.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
					return nil, errors.New("dial boom")
				})

				err := client.do(context.Background(), http.MethodGet, "/x", nil, nil)

				must.ErrorIs(err, domain.ErrKafkaConnectUnavailable)
				must.NotErrorIs(err, context.Canceled)
			},
		},
		{
			name: "read body error",
			run: func(t *testing.T) {
				t.Helper()

				must := require.New(t)

				client := NewHTTPClient(Config{
					BaseURL:           "http://example.com",
					Timeout:           time.Second,
					RetryMaxAttempts:  3,
					RetryInitialDelay: time.Millisecond,
				}, plugins.NewRegistry(plugins.PostgresAdapter{}))
				client.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       errReader{err: errors.New("read boom")},
						Header:     make(http.Header),
					}, nil
				})

				err := client.do(context.Background(), http.MethodGet, "/x", nil, &struct{}{})

				must.ErrorContains(err, "read body")
			},
		},
		{
			name: "unmarshal error",
			run: func(t *testing.T) {
				t.Helper()

				must := require.New(t)

				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					_, _ = w.Write([]byte(`not-json`)) //nolint:errcheck // test response write
				}))
				t.Cleanup(srv.Close)

				client := NewHTTPClient(Config{
					BaseURL:           srv.URL,
					Timeout:           time.Second,
					RetryMaxAttempts:  3,
					RetryInitialDelay: time.Millisecond,
				}, plugins.NewRegistry(plugins.PostgresAdapter{}))

				var out struct {
					Name string `json:"name"`
				}

				err := client.do(context.Background(), http.MethodGet, "/x", nil, &out)

				must.ErrorContains(err, "unmarshal response")
			},
		},
		{
			name: "500 maps to unavailable",
			run: func(t *testing.T) {
				t.Helper()

				must := require.New(t)

				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`boom`)) //nolint:errcheck // test response write
				}))
				t.Cleanup(srv.Close)

				client := NewHTTPClient(Config{
					BaseURL:           srv.URL,
					Timeout:           time.Second,
					RetryMaxAttempts:  3,
					RetryInitialDelay: time.Millisecond,
				}, plugins.NewRegistry(plugins.PostgresAdapter{}))

				err := client.do(context.Background(), http.MethodGet, "/x", nil, nil)

				must.ErrorIs(err, domain.ErrKafkaConnectUnavailable)

				var herr *HTTPError

				must.ErrorAs(err, &herr)
				must.Equal(500, herr.StatusCode)
				must.Equal("boom", herr.Body)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.run(t)
		})
	}
}

func TestHTTPClient_WithRetry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "success on first attempt",
			run: func(t *testing.T) {
				t.Helper()

				must := require.New(t)

				client := NewHTTPClient(Config{
					BaseURL:           "http://example.com",
					Timeout:           time.Second,
					RetryMaxAttempts:  3,
					RetryInitialDelay: time.Millisecond,
				}, plugins.NewRegistry(plugins.PostgresAdapter{}))

				calls := 0

				err := client.withRetry(context.Background(), func(error) bool { return true }, func() error {
					calls++

					return nil
				})

				must.NoError(err)
				must.Equal(1, calls)
			},
		},
		{
			name: "non retryable returns immediately",
			run: func(t *testing.T) {
				t.Helper()

				must := require.New(t)

				client := NewHTTPClient(Config{
					BaseURL:           "http://example.com",
					Timeout:           time.Second,
					RetryMaxAttempts:  3,
					RetryInitialDelay: time.Millisecond,
				}, plugins.NewRegistry(plugins.PostgresAdapter{}))

				calls := 0
				sentinel := errors.New("fatal boom")

				err := client.withRetry(context.Background(), func(error) bool { return false }, func() error {
					calls++

					return sentinel
				})

				must.ErrorIs(err, sentinel)
				must.Equal(1, calls)
			},
		},
		{
			name: "canceled context aborts wait",
			run: func(t *testing.T) {
				t.Helper()

				must := require.New(t)

				client := NewHTTPClient(Config{
					BaseURL:           "http://example.com",
					Timeout:           time.Second,
					RetryMaxAttempts:  3,
					RetryInitialDelay: time.Hour,
				}, plugins.NewRegistry(plugins.PostgresAdapter{}))

				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				calls := 0

				err := client.withRetry(ctx, func(error) bool { return true }, func() error {
					calls++

					return errors.New("retryable boom")
				})

				must.ErrorIs(err, context.Canceled)
				must.Equal(1, calls)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.run(t)
		})
	}
}

func TestHTTPClient_RetryWait(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		initialDelay time.Duration
		attempt      int
		wantExact    time.Duration
		wantMin      time.Duration
		wantMax      time.Duration
		exact        bool
	}{
		{
			name:         "tiny delay returns base",
			initialDelay: time.Duration(1),
			attempt:      1,
			wantExact:    time.Duration(1),
			exact:        true,
		},
		{
			name:         "small delay returns base",
			initialDelay: time.Duration(2),
			attempt:      1,
			wantExact:    time.Duration(2),
			exact:        true,
		},
		{
			name:         "normal delay within jitter",
			initialDelay: 100 * time.Millisecond,
			attempt:      2,
			wantMin:      160 * time.Millisecond,
			wantMax:      240 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			client := NewHTTPClient(Config{
				BaseURL:           "http://example.com",
				Timeout:           time.Second,
				RetryMaxAttempts:  5,
				RetryInitialDelay: tt.initialDelay,
			}, plugins.NewRegistry(plugins.PostgresAdapter{}))

			got := client.retryWait(tt.attempt)

			if tt.exact {
				must.Equal(tt.wantExact, got)

				return
			}

			must.GreaterOrEqual(got, tt.wantMin)
			must.LessOrEqual(got, tt.wantMax)
		})
	}
}

func TestBodyOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "http error returns body",
			err:  &HTTPError{StatusCode: http.StatusNotFound, Body: "not found"},
			want: "not found",
		},
		{
			name: "wrapped http error returns body",
			err:  fmt.Errorf("wrap: %w", &HTTPError{StatusCode: http.StatusInternalServerError, Body: "boom"}),
			want: "boom",
		},
		{
			name: "generic error returns empty",
			err:  errors.New("boom"),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			must.Equal(tt.want, bodyOf(tt.err))
		})
	}
}
