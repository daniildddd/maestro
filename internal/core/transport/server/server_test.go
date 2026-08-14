package server_test

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"go.uber.org/goleak"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/stretchr/testify/require"

	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/server"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func nopLogger() *core_logger.Logger {
	return &core_logger.Logger{Logger: zap.NewNop()}
}

func newObservableLogger(t *testing.T, lvl zapcore.Level) (*core_logger.Logger, *observer.ObservedLogs) {
	t.Helper()

	core, recorder := observer.New(lvl)

	return &core_logger.Logger{Logger: zap.New(core)}, recorder
}

func freeAddr(t *testing.T) string {
	t.Helper()
	must := require.New(t)

	listener, err := net.Listen("tcp", "localhost:0")
	must.NoError(err)

	addr := listener.Addr().String()
	must.NoError(listener.Close())

	return addr
}

func runAndAwaitListening(
	ctx context.Context,
	t *testing.T,
	srv *server.HTTPServer,
	addr string,
) <-chan error {
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
	}, 2*time.Second, 50*time.Millisecond, "server should start listening")

	return errCh
}

func awaitRunResult(t *testing.T, errCh <-chan error) error {
	t.Helper()

	select {
	case err := <-errCh:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after context cancellation")

		return nil
	}
}

func TestNewHTTPServer(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil server", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		cfg := server.Config{Addr: "localhost:0"}
		log := nopLogger()

		srv := server.NewHTTPServer(cfg, log)

		must.NotNil(srv)
	})
}

//nolint:paralleltest // binds real OS ports; cannot run concurrently with other Run tests
func TestHTTPServer_Run(t *testing.T) {
	t.Run("shuts down gracefully on context cancellation", func(t *testing.T) {
		must := require.New(t)

		addr := freeAddr(t)

		log, recordedLogs := newObservableLogger(t, zapcore.InfoLevel)
		srv := server.NewHTTPServer(
			server.Config{Addr: addr, ShutdownTimeout: 5 * time.Second},
			log,
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		errCh := runAndAwaitListening(ctx, t, srv, addr)

		cancel()

		must.NoError(awaitRunResult(t, errCh))

		logs := recordedLogs.All()
		must.Len(logs, 1)
		must.Equal(zapcore.InfoLevel, logs[0].Level)
		must.Equal("starting graceful shutdown", logs[0].Message)

		ctxMap := logs[0].ContextMap()
		must.Equal(5*time.Second, ctxMap["timeout"])
	})

	t.Run("returns error when address already in use", func(t *testing.T) {
		must := require.New(t)

		listener, err := net.Listen("tcp", "localhost:0")
		must.NoError(err)

		defer func() { _ = listener.Close() }() //nolint:errcheck // test-only

		addr := listener.Addr().String()

		srv := server.NewHTTPServer(server.Config{Addr: addr}, nopLogger())

		err = srv.Run(context.Background())

		must.Error(err)
		must.Contains(err.Error(), "listen and serve HTTP")
	})

	t.Run("returns error when graceful shutdown times out", func(t *testing.T) {
		must := require.New(t)

		addr := freeAddr(t)

		log, recordedLogs := newObservableLogger(t, zapcore.InfoLevel)

		hold := make(chan struct{})
		started := make(chan struct{})

		defer close(hold)

		srv := server.NewHTTPServer(
			server.Config{
				Addr:            addr,
				ShutdownTimeout: 1 * time.Millisecond,
			},
			log,
		)

		srv.RegisterRoute(server.Route{
			Method: http.MethodGet,
			Path:   "/slow",
			Handler: func(_ http.ResponseWriter, _ *http.Request) {
				close(started)
				<-hold
			},
		})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		errCh := runAndAwaitListening(ctx, t, srv, addr)

		go func() {
			_, _ = http.Get("http://" + addr + "/slow") //nolint:errcheck,bodyclose // test-only
		}()

		<-started

		cancel()

		err := awaitRunResult(t, errCh)
		must.Error(err)
		must.Contains(err.Error(), "graceful shutdown")
		must.ErrorIs(err, context.DeadlineExceeded)

		logs := recordedLogs.All()
		must.Len(logs, 1)
		must.Equal(zapcore.InfoLevel, logs[0].Level)
		must.Equal("starting graceful shutdown", logs[0].Message)

		ctxMap := logs[0].ContextMap()
		must.Equal(1*time.Millisecond, ctxMap["timeout"])
	})
}
