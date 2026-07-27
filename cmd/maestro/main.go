package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT|syscall.SIGTERM,
	)
	defer cancel()

	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init logger:", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info("logger successful init")

	postgresPool, err := pgxadapter.NewPool(
		ctx,
		pgxadapter.NewConfigMust(),
	)
	if err != nil {
		panic(fmt.Errorf("create postgres adapter pool: %v", err))
	}
	defer postgresPool.Close()

	authRepository := repository.NewAuthRepository(postgresPool)

	bcryptHasher, err := hasher.NewBcryptHasher(hasher.NewConfigMust())
	if err != nil {
		panic(fmt.Errorf("init bcrypt hasher: %v", err))
	}

	accessManager := access.NewManager(access.NewConfigMust())
	refreshManager, err := refresh.NewManager(refresh.NewConfigMust())
	if err != nil {
		panic(fmt.Errorf("init refresh token manager: %v", err))
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
		middleware.Logger(logger),
		middleware.RequestID(),
		middleware.Trace(),
		middleware.Recovery(),
	}

	publicV1 := server.NewAPIVersionRouter(
		authTransportHTTP.PublicRoutes(),
		server.ApiVersion1,
		baseMW...,
	)

	srv := server.NewHTTPServer(
		server.NewConfigMust(),
		logger,
	)

	srv.RegisterAPIRouters(publicV1)

	if err = srv.Run(ctx); err != nil {
		logger.Fatal("HTTP server run error", zap.Error(err))
	}
}
