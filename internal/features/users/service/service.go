package service

import (
	"context"

	"github.com/google/uuid"

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

	GetUserByID(
		ctx context.Context,
		id uuid.UUID,
	) (domain.User, error)

	DeleteUser(
		ctx context.Context,
		id uuid.UUID,
	) error
}
