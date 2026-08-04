package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/middleware"
)

type HTTPServer struct {
	mux         *http.ServeMux
	config      Config
	log         *core_logger.Logger
	middlewares []middleware.Middleware
}

func NewHTTPServer(
	config Config,
	log *core_logger.Logger,
) *HTTPServer {
	return &HTTPServer{
		mux:    http.NewServeMux(),
		config: config,
		log:    log,
	}
}

func (s *HTTPServer) RegisterAPIRouters(routers ...*APIVersionRouter) {
	for _, v := range routers {
		handlers := v.Handlers()

		for pattern, handler := range handlers {
			s.mux.Handle(pattern, handler)
		}
	}
}

func (s *HTTPServer) RegisterRoute(route ...Route) {
	for _, v := range route {
		pattern := v.Method + " " + v.Path
		s.mux.Handle(pattern, v.WithMiddleware())
	}
}

func (s *HTTPServer) Run(ctx context.Context) error {
	mux := middleware.ChainMiddleware(s.mux, s.middlewares...)

	server := http.Server{
		Addr:                         s.config.Addr,
		Handler:                      mux,
		DisableGeneralOptionsHandler: s.config.DisableGeneralOptionsHandler,
		ReadTimeout:                  s.config.ReadTimeout,
		ReadHeaderTimeout:            s.config.ReadHeaderTimeout,
		WriteTimeout:                 s.config.WriteTimeout,
		IdleTimeout:                  s.config.IdleTimeout,
		MaxHeaderBytes:               s.config.MaxHeaderBytes,
	}

	ch := make(chan error)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				ch <- err
			}
		}

		close(ch)
	}()

	select {
	case <-ctx.Done():
		s.log.Debug(
			"starting graceful shutdown",
			zap.Duration("timeout", s.config.ShutdownTimeout),
		)

		ctxShutdown, cancelShutdown := context.WithTimeout(
			context.Background(),
			s.config.ShutdownTimeout,
		)
		defer cancelShutdown()

		if err := server.Shutdown(ctxShutdown); err != nil { //nolint:contextcheck // graceful shutdown needs a fresh root context: parent ctx is already cancelled here (case <-ctx.Done())
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}

			if closeErr := server.Close(); closeErr != nil {
				return errors.Join(
					fmt.Errorf("graceful shutdown: %w", err),
					fmt.Errorf("force close: %w", closeErr),
				)
			}

			return fmt.Errorf("graceful shutdown: %w", err)
		}
	case err := <-ch:
		return fmt.Errorf("listen and serve HTTP: %w", err)
	}

	return nil
}
