package repository

import (
	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
)

type AuditRepository struct {
	pool core_postgres_pool.Pool
}

func NewAuditRepository(
	pool core_postgres_pool.Pool,
) *AuditRepository {
	return &AuditRepository{
		pool: pool,
	}
}
