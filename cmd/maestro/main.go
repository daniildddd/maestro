package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/repository/postgres/pgxadapter"
	"github.com/daniildddd/maestro/internal/core/security/access"
	"github.com/daniildddd/maestro/internal/core/security/hasher"
	"github.com/daniildddd/maestro/internal/core/security/refresh"
	"github.com/daniildddd/maestro/internal/core/transport/middleware"
	"github.com/daniildddd/maestro/internal/core/transport/server"
	"github.com/daniildddd/maestro/internal/features/auth/repository"
	"github.com/daniildddd/maestro/internal/features/auth/service"
	"github.com/daniildddd/maestro/internal/features/auth/transport"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		slog.New(slog.NewTextHandler(os.Stderr, nil)).
			Error("init logger", slog.String("error", err.Error()))

		return 1
	}

	defer func() {
		if err := logger.Close(); err != nil {
			slog.New(slog.NewTextHandler(os.Stderr, nil)).
				Error("close logger gracefully", slog.String("error", err.Error()))
		}
	}()

	logger.Info("logger successful init")

	postgresPool, err := pgxadapter.NewPool(
		ctx,
		pgxadapter.NewConfigMust(),
	)
	if err != nil {
		logger.Error("create postgres adapter pool", zap.Error(err))

		return 1
	}
	defer postgresPool.Close()

	authRepository := repository.NewAuthRepository(postgresPool)

	bcryptHasher, err := hasher.NewBcryptHasher(hasher.NewConfigMust())
	if err != nil {
		logger.Error("init bcrypt hasher", zap.Error(err))

		return 1
	}

	accessManager := access.NewManager(access.NewConfigMust())

	refreshManager, err := refresh.NewManager(refresh.NewConfigMust())
	if err != nil {
		logger.Error("init refresh token manager", zap.Error(err))

		return 1
	}

	authService := service.NewAuthService(
		authRepository,
		bcryptHasher,
		accessManager,
		refreshManager,
	)

	authTransportHTTP := transport.NewAuthHTTPHandler(
		authService,
		transport.NewConfigMust(),
	)

	cfgMiddleware := middleware.NewConfigMust()

	baseMW := []middleware.Middleware{
		middleware.CORS(cfgMiddleware.AllowedOrigins),
		middleware.RequestID(),
		middleware.Logger(logger),
		middleware.Trace(),
		middleware.Recovery(),
	}

	publicV1 := server.NewAPIVersionRouter(
		authTransportHTTP.PublicRoutes(),
		server.ApiVersion1,
		baseMW...,
	)

	privateMW := make([]middleware.Middleware, len(baseMW), len(baseMW)+1)
	copy(privateMW, baseMW)

	privateMW = append(privateMW, middleware.Auth(accessManager))

	privateV1 := server.NewAPIVersionRouter(
		authTransportHTTP.PrivateRoutes(),
		server.ApiVersion1,
		privateMW...,
	)

	srv := server.NewHTTPServer(
		server.NewConfigMust(),
		logger,
	)

	srv.RegisterAPIRouters(publicV1, privateV1)

	if err = srv.Run(ctx); err != nil {
		logger.Error("HTTP server stopped", zap.Error(err))

		return 1
	}

	logger.Info("HTTP server shut down gracefully")

	return 0
}
