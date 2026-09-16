package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	core_metrics "github.com/daniildddd/maestro/internal/core/metrics"
	"github.com/daniildddd/maestro/internal/core/repository/postgres/pgxadapter"
	"github.com/daniildddd/maestro/internal/core/security/access"
	"github.com/daniildddd/maestro/internal/core/security/hasher"
	"github.com/daniildddd/maestro/internal/core/security/refresh"
	"github.com/daniildddd/maestro/internal/core/transport/health"
	"github.com/daniildddd/maestro/internal/core/transport/middleware"
	"github.com/daniildddd/maestro/internal/core/transport/server"
	auditrepository "github.com/daniildddd/maestro/internal/features/audit/repository"
	auditservice "github.com/daniildddd/maestro/internal/features/audit/service"
	audittransport "github.com/daniildddd/maestro/internal/features/audit/transport"
	"github.com/daniildddd/maestro/internal/features/auth/cleanup"
	"github.com/daniildddd/maestro/internal/features/auth/repository"
	"github.com/daniildddd/maestro/internal/features/auth/service"
	"github.com/daniildddd/maestro/internal/features/auth/transport"
	"github.com/daniildddd/maestro/internal/features/connectors/collector"
	"github.com/daniildddd/maestro/internal/features/connectors/dbcheck"
	"github.com/daniildddd/maestro/internal/features/connectors/kafkaconnect"
	"github.com/daniildddd/maestro/internal/features/connectors/plugins"
	connectorsservice "github.com/daniildddd/maestro/internal/features/connectors/service"
	connectorstransport "github.com/daniildddd/maestro/internal/features/connectors/transport"
	usersrepository "github.com/daniildddd/maestro/internal/features/users/repository"
	usersservice "github.com/daniildddd/maestro/internal/features/users/service"
	userstransport "github.com/daniildddd/maestro/internal/features/users/transport"
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

	auditRepository := auditrepository.NewAuditRepository(postgresPool)

	auditService := auditservice.NewAuditService(auditRepository)

	auditTransportHTTP := audittransport.NewAuditHTTPHandler(auditService)

	refreshTokenCleanup := cleanup.NewWorker(
		authRepository,
		logger,
		cleanup.NewConfigMust(),
	)

	go refreshTokenCleanup.Run(ctx)

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
		auditRepository,
	)

	authTransportHTTP := transport.NewAuthHTTPHandler(
		authService,
		transport.NewConfigMust(),
	)

	usersRepository := usersrepository.NewUsersRepository(postgresPool)

	usersService := usersservice.NewUsersService(
		usersRepository,
		bcryptHasher,
		auditRepository,
	)

	usersTransportHTTP := userstransport.NewUsersHTTPHandler(usersService)

	pluginRegistry := plugins.NewRegistry(
		plugins.PostgresAdapter{},
		plugins.MySQLAdapter{},
		plugins.MariaDBAdapter{},
		plugins.MongoDBAdapter{},
		plugins.SQLServerAdapter{},
		plugins.OracleAdapter{},
		plugins.Db2Adapter{},
		plugins.InformixAdapter{},
		plugins.VitessAdapter{},
	)

	kafkaConnectClient := kafkaconnect.NewHTTPClient(
		kafkaconnect.NewConfigMust(),
		pluginRegistry,
	)

	dbCheckConfig := dbcheck.NewConfigMust()

	connectorsService := connectorsservice.NewConnectorsService(
		kafkaConnectClient,
		auditRepository,
		dbcheck.NewPostgresChecker(dbCheckConfig),
	)

	connectorsTransportHTTP := connectorstransport.NewConnectorsHTTPHandler(connectorsService)

	cfgMiddleware := middleware.NewConfigMust()

	metricsCfg := core_metrics.NewConfigMust()
	appMetrics := core_metrics.New(metricsCfg)

	baseMW := []middleware.Middleware{
		middleware.Metrics(appMetrics),
		middleware.CORS(cfgMiddleware.AllowedOrigins),
		middleware.RequestID(),
		middleware.ClientInfo(),
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

	privateRoutes := append(
		authTransportHTTP.PrivateRoutes(),
		usersTransportHTTP.PrivateRoutes()...,
	)
	privateRoutes = append(
		privateRoutes,
		connectorsTransportHTTP.PrivateRoutes()...,
	)
	privateRoutes = append(
		privateRoutes,
		auditTransportHTTP.PrivateRoutes()...,
	)

	privateV1 := server.NewAPIVersionRouter(
		privateRoutes,
		server.ApiVersion1,
		privateMW...,
	)

	srv := server.NewHTTPServer(
		server.NewConfigMust(),
		logger,
	)

	srv.RegisterAPIRouters(publicV1, privateV1)
	srv.RegisterNotFound(logger)

	healthHandler := health.NewHealthHandler(postgresPool, kafkaConnectClient, health.NewConfigMust())
	srv.RegisterRoute(healthHandler.Routes()...)

	metricsSrv := core_metrics.NewServer(appMetrics, metricsCfg, logger)

	metricsCollector := collector.NewCollector(
		kafkaConnectClient,
		appMetrics,
		logger,
		collector.NewCollectorConfigMust(),
	)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := srv.Run(ctx); err != nil {
			return fmt.Errorf("http server: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		if err := metricsSrv.Run(ctx); err != nil {
			return fmt.Errorf("metrics server: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		metricsCollector.Run(ctx)

		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Error("application stopped", zap.Error(err))

		return 1
	}

	logger.Info("application shut down gracefully")

	return 0
}
