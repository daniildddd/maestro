package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
)

type UsersService struct {
	usersRepository UsersRepository
	passwordHasher  PasswordHasher
}

func NewUsersService(
	usersRepository UsersRepository,
	passwordHasher PasswordHasher,
) *UsersService {
	return &UsersService{
		usersRepository: usersRepository,
		passwordHasher:  passwordHasher,
	}
}

type PasswordHasher interface {
	Verify(
		hash string,
		plain string,
	) error

	Hash(
		password string,
	) (string, error)
}

type UsersRepository interface {
	GetUsers(
		ctx context.Context,
		filter domain.UserFilter,
	) ([]domain.User, error)

	CreateUser(
		ctx context.Context,
		user domain.User,
	) (domain.User, error)

	GetUserByID(
		ctx context.Context,
		id uuid.UUID,
	) (domain.User, error)

	DeleteUser(
		ctx context.Context,
		id uuid.UUID,
	) error

	ChangePassword(
		ctx context.Context,
		id uuid.UUID,
		passwordHash string,
	) error

	UpdateUser(
		ctx context.Context,
		id uuid.UUID,
		username string,
	) (domain.User, error)
}
