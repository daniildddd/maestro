package repository

import (
	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
)

type AuditRepository struct {
	pool   core_postgres_pool.Pool
	logger *core_logger.Logger
}

func NewAuditRepository(
	pool core_postgres_pool.Pool,
	log *core_logger.Logger,
) *AuditRepository {
	return &AuditRepository{
		pool:   pool,
		logger: log,
	}
}
