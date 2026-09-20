package metrics_test

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/metrics"
)

func metricsNopLogger() *logger.Logger {
	return &logger.Logger{Logger: zap.NewNop()}
}

func metricsFreeAddr(t *testing.T) string {
	t.Helper()

	must := require.New(t)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	must.NoError(err)

	addr := listener.Addr().String()
	must.NoError(listener.Close())

	return addr
}

func metricsRunAndAwait(ctx context.Context, t *testing.T, srv *metrics.Server, addr string) <-chan error {
	t.Helper()

	errCh := make(chan error, 1)

	go func() {
		errCh <- srv.Run(ctx)
	}()

	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err != nil {
			return false
		}

		_ = conn.Close() //nolint:errcheck // test-only: closing probe connection

		return true
	}, 10*time.Second, 50*time.Millisecond, "metrics server should start listening")

	return errCh
}

func metricsAwaitRun(t *testing.T, errCh <-chan error) error {
	t.Helper()

	select {
	case err := <-errCh:
		return err
	case <-time.After(10 * time.Second):
		t.Fatal("metrics Run did not return after context cancellation")

		return nil
	}
}

func TestNewServer(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	m := metrics.New(metrics.Config{Buckets: []float64{0.1, 1}})

	srv := metrics.NewServer(m, metrics.Config{
		Addr:              "127.0.0.1:0",
		ReadHeaderTimeout: 5 * time.Second,
		ShutdownTimeout:   5 * time.Second,
	}, metricsNopLogger())

	must.NotNil(srv)
}

//nolint:paralleltest // binds real OS ports; cannot run concurrently with other Run tests
func TestServer_Run(t *testing.T) {
	t.Run("shuts down gracefully on context cancellation", func(t *testing.T) {
		must := require.New(t)

		addr := metricsFreeAddr(t)

		m := metrics.New(metrics.Config{Buckets: []float64{0.1, 1}})

		srv := metrics.NewServer(m, metrics.Config{
			Addr:              addr,
			ReadHeaderTimeout: 5 * time.Second,
			ShutdownTimeout:   5 * time.Second,
		}, metricsNopLogger())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		errCh := metricsRunAndAwait(ctx, t, srv, addr)

		resp, err := http.Get("http://" + addr + "/metrics")
		must.NoError(err)
		must.Equal(http.StatusOK, resp.StatusCode)
		_ = resp.Body.Close() //nolint:errcheck // test-only

		cancel()
		must.NoError(metricsAwaitRun(t, errCh))
	})

	t.Run("returns error when address already in use", func(t *testing.T) {
		must := require.New(t)

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		must.NoError(err)

		defer func() { _ = listener.Close() }() //nolint:errcheck // test-only

		addr := listener.Addr().String()

		m := metrics.New(metrics.Config{Buckets: []float64{0.1, 1}})

		srv := metrics.NewServer(m, metrics.Config{
			Addr:              addr,
			ReadHeaderTimeout: 5 * time.Second,
			ShutdownTimeout:   5 * time.Second,
		}, metricsNopLogger())

		err = srv.Run(context.Background())
		must.Error(err)
		must.Contains(err.Error(), "listen and serve metrics")
	})
}
