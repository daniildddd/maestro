package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/daniildddd/maestro/internal/core/logger"
)

type Server struct {
	srv             *http.Server
	log             *logger.Logger
	shutdownTimeout time.Duration
}

func NewServer(m *Metrics, cfg Config, log *logger.Logger) *Server {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", m.Handler())

	return &Server{
		srv: &http.Server{
			Addr:              cfg.Addr,
			Handler:           mux,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		},
		log:             log,
		shutdownTimeout: cfg.ShutdownTimeout,
	}
}

func (s *Server) Run(ctx context.Context) error {
	const op = "metrics.server.Run"

	ch := make(chan error, 1)

	go func() {
		if err := s.srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				ch <- err
			}
		}

		close(ch)
	}()

	select {
	case <-ctx.Done():
		s.log.Info(
			"starting metrics graceful shutdown",
			zap.Duration("timeout", s.shutdownTimeout),
		)

		ctxShutdown, cancelShutdown := context.WithTimeout(
			context.Background(),
			s.shutdownTimeout,
		)
		defer cancelShutdown()

		//nolint:contextcheck // graceful shutdown needs a fresh root context:
		if err := s.srv.Shutdown(ctxShutdown); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}

			if closeErr := s.srv.Close(); closeErr != nil {
				return errors.Join(
					fmt.Errorf("%s: graceful shutdown: %w", op, err),
					fmt.Errorf("%s: force close: %w", op, closeErr),
				)
			}

			return fmt.Errorf("%s: graceful shutdown: %w", op, err)
		}
	case err := <-ch:
		return fmt.Errorf("%s: listen and serve metrics: %w", op, err)
	}

	return nil
}
