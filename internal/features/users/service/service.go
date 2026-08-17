package service

import (
	"context"

	"github.com/daniildddd/maestro/internal/core/domain"
)

type UsersService struct {
	usersRepository UsersRepository
}

func NewUsersService(
	usersRepository UsersRepository,
) *UsersService {
	return &UsersService{
		usersRepository: usersRepository,
	}
}

type UsersRepository interface {
	GetUsers(
		ctx context.Context,
		filter domain.UserFilter,
	) ([]domain.User, error)
}
