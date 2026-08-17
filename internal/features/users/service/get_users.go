package service

import (
	"context"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func (s *UsersService) GetUsers(
	ctx context.Context,
	filter domain.UserFilter,
) ([]domain.User, error) {
	const op = "users.service.GetUsers"

	users, err := s.usersRepository.GetUsers(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}
