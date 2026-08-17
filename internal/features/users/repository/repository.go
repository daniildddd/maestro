package repository

import (
	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
)

type UsersRepository struct {
	pool core_postgres_pool.Pool
}

func NewUsersRepository(
	pool core_postgres_pool.Pool,
) *UsersRepository {
	return &UsersRepository{
		pool: pool,
	}
}
