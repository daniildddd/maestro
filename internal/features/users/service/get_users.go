package service

import (
	"context"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func (s *UsersService) GetUsers(
	ctx context.Context,
	filter *domain.UserFilter,
) ([]domain.User, bool, error) {
	const op = "users.service.GetUsers"

	users, err := s.usersRepository.GetUsers(ctx, filter)
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", op, err)
	}

	hasMore := len(users) > filter.Limit
	if hasMore {
		trimmed := make([]domain.User, 0, filter.Limit)
		trimmed = append(trimmed, users[:filter.Limit]...)

		users = trimmed
	}

	return users, hasMore, nil
}
