package repository

import (
	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
)

type AuthRepository struct {
	pool core_postgres_pool.Pool
}

func NewAuthRepository(
	pool core_postgres_pool.Pool,
) *AuthRepository {
	return &AuthRepository{
		pool: pool,
	}
}
