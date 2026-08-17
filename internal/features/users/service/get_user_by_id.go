package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func (s *UsersService) GetUserByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.User, error) {
	const op = "users.service.GetUserByID"

	user, err := s.usersRepository.GetUserByID(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
