package kafkaconnect

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gotest "github.com/allure-framework/allure-go/commons/gotest" //nolint:revive // package clause is allure, gotest keeps call sites readable
	"github.com/allure-framework/allure-go/commons/httpexchange"
	allurerequire "github.com/allure-framework/allure-go/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/features/connectors/plugins"
)

func TestNewHTTPClient(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "new http client", func(a *gotest.Context) {
		must := allurerequire.New(a)

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
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("HTTP client"),
		gotest.WithSeverity("minor"),
		gotest.WithOwner("qa-intern"),
	)
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

	gotest.Test(t, "http client ping", func(a *gotest.Context) {
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
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

				var gotPath string

				var gotMethod string

				srv := httptest.NewServer(httpexchange.NewHandler(a.Context(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					gotPath = r.URL.Path
					gotMethod = r.Method

					w.WriteHeader(tt.status)
				})))
				a.T().Cleanup(srv.Close)

				client := NewHTTPClient(Config{
					BaseURL:           srv.URL,
					Timeout:           time.Second,
					RetryMaxAttempts:  3,
					RetryInitialDelay: time.Millisecond,
				}, plugins.NewRegistry(plugins.PostgresAdapter{}))

				err := client.Ping(a.Context())

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
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("HTTP client"),
		gotest.WithSeverity("normal"),
		gotest.WithOwner("qa-intern"),
	)
}

func TestHTTPClient_Do(t *testing.T) { //nolint:maintidx // table volume comes from per-case servers
	t.Parallel()

	gotest.Test(t, "http client do", func(a *gotest.Context) {
		tests := []struct {
			name string
			run  func(a *gotest.Context)
		}{
			{
				name: "200 unmarshals response",
				run: func(a *gotest.Context) {
					a.T().Helper()

					must := allurerequire.New(a)

					srv := httptest.NewServer(httpexchange.NewHandler(a.Context(), http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(`{"name":"pg-connector"}`)) //nolint:errcheck // test response write
					})))
					a.T().Cleanup(srv.Close)

					client := NewHTTPClient(Config{
						BaseURL:           srv.URL,
						Timeout:           time.Second,
						RetryMaxAttempts:  3,
						RetryInitialDelay: time.Millisecond,
					}, plugins.NewRegistry(plugins.PostgresAdapter{}))

					var got struct {
						Name string `json:"name"`
					}

					err := client.do(a.Context(), http.MethodGet, "/x", nil, &got)

					must.NoError(err)
					must.Equal(struct {
						Name string `json:"name"`
					}{Name: "pg-connector"}, got)
				},
			},
			{
				name: "marshal error",
				run: func(a *gotest.Context) {
					a.T().Helper()

					must := allurerequire.New(a)

					client := NewHTTPClient(Config{
						BaseURL:           "http://example.com",
						Timeout:           time.Second,
						RetryMaxAttempts:  3,
						RetryInitialDelay: time.Millisecond,
					}, plugins.NewRegistry(plugins.PostgresAdapter{}))

					err := client.do(a.Context(), http.MethodPost, "/x", func() {}, nil)

					must.ErrorContains(err, "marshal body")
				},
			},
			{
				name: "build request error",
				run: func(a *gotest.Context) {
					a.T().Helper()

					must := allurerequire.New(a)

					client := NewHTTPClient(Config{
						BaseURL:           "://bad-url",
						Timeout:           time.Second,
						RetryMaxAttempts:  3,
						RetryInitialDelay: time.Millisecond,
					}, plugins.NewRegistry(plugins.PostgresAdapter{}))

					err := client.do(a.Context(), http.MethodGet, "/x", nil, nil)

					must.ErrorContains(err, "build request")
				},
			},
			{
				name: "transport error maps to unavailable",
				run: func(a *gotest.Context) {
					a.T().Helper()

					must := allurerequire.New(a)

					client := NewHTTPClient(Config{
						BaseURL:           "http://example.com",
						Timeout:           time.Second,
						RetryMaxAttempts:  3,
						RetryInitialDelay: time.Millisecond,
					}, plugins.NewRegistry(plugins.PostgresAdapter{}))
					client.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
						return nil, errors.New("dial boom")
					})

					err := client.do(a.Context(), http.MethodGet, "/x", nil, nil)

					must.ErrorIs(err, domain.ErrKafkaConnectUnavailable)
					must.NotErrorIs(err, context.Canceled)
				},
			},
			{
				name: "read body error",
				run: func(a *gotest.Context) {
					a.T().Helper()

					must := allurerequire.New(a)

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

					err := client.do(a.Context(), http.MethodGet, "/x", nil, &struct{}{})

					must.ErrorContains(err, "read body")
				},
			},
			{
				name: "unmarshal error",
				run: func(a *gotest.Context) {
					a.T().Helper()

					must := allurerequire.New(a)

					srv := httptest.NewServer(httpexchange.NewHandler(a.Context(), http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						_, _ = w.Write([]byte(`not-json`)) //nolint:errcheck // test response write
					})))
					a.T().Cleanup(srv.Close)

					client := NewHTTPClient(Config{
						BaseURL:           srv.URL,
						Timeout:           time.Second,
						RetryMaxAttempts:  3,
						RetryInitialDelay: time.Millisecond,
					}, plugins.NewRegistry(plugins.PostgresAdapter{}))

					var out struct {
						Name string `json:"name"`
					}

					err := client.do(a.Context(), http.MethodGet, "/x", nil, &out)

					must.ErrorContains(err, "unmarshal response")
				},
			},
			{
				name: "500 maps to unavailable",
				run: func(a *gotest.Context) {
					a.T().Helper()

					must := allurerequire.New(a)

					srv := httptest.NewServer(httpexchange.NewHandler(a.Context(), http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusInternalServerError)
						_, _ = w.Write([]byte(`boom`)) //nolint:errcheck // test response write
					})))
					a.T().Cleanup(srv.Close)

					client := NewHTTPClient(Config{
						BaseURL:           srv.URL,
						Timeout:           time.Second,
						RetryMaxAttempts:  3,
						RetryInitialDelay: time.Millisecond,
					}, plugins.NewRegistry(plugins.PostgresAdapter{}))

					err := client.do(a.Context(), http.MethodGet, "/x", nil, nil)

					must.ErrorIs(err, domain.ErrKafkaConnectUnavailable)

					var herr *HTTPError

					must.ErrorAs(err, &herr)
					must.Equal(500, herr.StatusCode)
					must.Equal("boom", herr.Body)
				},
			},
			{
				name: "client timeout aborts slow server",
				run: func(a *gotest.Context) {
					a.T().Helper()

					must := allurerequire.New(a)

					srv := httptest.NewServer(httpexchange.NewHandler(a.Context(), http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						time.Sleep(200 * time.Millisecond)

						_, _ = w.Write([]byte(`{}`)) //nolint:errcheck // test response write
					})))
					a.T().Cleanup(srv.Close)

					client := NewHTTPClient(Config{
						BaseURL:           srv.URL,
						Timeout:           50 * time.Millisecond,
						RetryMaxAttempts:  1,
						RetryInitialDelay: time.Millisecond,
					}, plugins.NewRegistry(plugins.PostgresAdapter{}))

					err := client.do(a.Context(), http.MethodGet, "/x", nil, nil)

					must.Error(err)
					must.ErrorIs(err, domain.ErrKafkaConnectUnavailable)
					must.ErrorContains(err, "Client.Timeout exceeded")
				},
			},
			{
				name: "generous timeout lets slow server answer",
				run: func(a *gotest.Context) {
					a.T().Helper()

					must := allurerequire.New(a)

					srv := httptest.NewServer(httpexchange.NewHandler(a.Context(), http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						time.Sleep(20 * time.Millisecond)

						_, _ = w.Write([]byte(`{"name":"pg-connector"}`)) //nolint:errcheck // test response write
					})))
					a.T().Cleanup(srv.Close)

					client := NewHTTPClient(Config{
						BaseURL:           srv.URL,
						Timeout:           5 * time.Second,
						RetryMaxAttempts:  1,
						RetryInitialDelay: time.Millisecond,
					}, plugins.NewRegistry(plugins.PostgresAdapter{}))

					var got struct {
						Name string `json:"name"`
					}

					err := client.do(a.Context(), http.MethodGet, "/x", nil, &got)

					must.NoError(err)
					must.Equal("pg-connector", got.Name)
				},
			},
		}

		for _, tt := range tests {
			a.Step(tt.name, func(a *gotest.Context) {
				tt.run(a)
			})
		}
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("HTTP client"),
		gotest.WithSeverity("minor"),
		gotest.WithOwner("qa-intern"),
	)
}

func TestHTTPClient_WithRetry(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "http client retry", func(a *gotest.Context) {
		tests := []struct {
			name string
			run  func(a *gotest.Context)
		}{
			{
				name: "success on first attempt",
				run: func(a *gotest.Context) {
					a.T().Helper()

					must := allurerequire.New(a)

					client := NewHTTPClient(Config{
						BaseURL:           "http://example.com",
						Timeout:           time.Second,
						RetryMaxAttempts:  3,
						RetryInitialDelay: time.Millisecond,
					}, plugins.NewRegistry(plugins.PostgresAdapter{}))

					calls := 0

					err := client.withRetry(a.Context(), func(error) bool { return true }, func() error {
						calls++

						return nil
					})

					must.NoError(err)
					must.Equal(1, calls)
				},
			},
			{
				name: "non retryable returns immediately",
				run: func(a *gotest.Context) {
					a.T().Helper()

					must := allurerequire.New(a)

					client := NewHTTPClient(Config{
						BaseURL:           "http://example.com",
						Timeout:           time.Second,
						RetryMaxAttempts:  3,
						RetryInitialDelay: time.Millisecond,
					}, plugins.NewRegistry(plugins.PostgresAdapter{}))

					calls := 0
					sentinel := errors.New("fatal boom")

					err := client.withRetry(a.Context(), func(error) bool { return false }, func() error {
						calls++

						return sentinel
					})

					must.ErrorIs(err, sentinel)
					must.Equal(1, calls)
				},
			},
			{
				name: "canceled context aborts wait",
				run: func(a *gotest.Context) {
					a.T().Helper()

					must := allurerequire.New(a)

					client := NewHTTPClient(Config{
						BaseURL:           "http://example.com",
						Timeout:           time.Second,
						RetryMaxAttempts:  3,
						RetryInitialDelay: time.Hour,
					}, plugins.NewRegistry(plugins.PostgresAdapter{}))

					ctx, cancel := context.WithCancel(a.Context())
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
			a.Step(tt.name, func(a *gotest.Context) {
				tt.run(a)
			})
		}
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("HTTP client"),
		gotest.WithSeverity("minor"),
		gotest.WithOwner("qa-intern"),
	)
}

func TestHTTPClient_RetryWait(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "http client retry wait", func(a *gotest.Context) {
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
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

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
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("HTTP client"),
		gotest.WithSeverity("minor"),
		gotest.WithOwner("qa-intern"),
	)
}

func TestBodyOf(t *testing.T) {
	t.Parallel()

	gotest.Test(t, "body of", func(a *gotest.Context) {
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
			a.Step(tt.name, func(a *gotest.Context) {
				must := allurerequire.New(a)

				must.Equal(tt.want, bodyOf(tt.err))
			})
		}
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("HTTP client"),
		gotest.WithSeverity("minor"),
		gotest.WithOwner("qa-intern"),
	)
}
