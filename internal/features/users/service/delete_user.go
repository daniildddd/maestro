package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *UsersService) DeleteUser(
	ctx context.Context,
	id uuid.UUID,
) error {
	const op = "users.service.DeleteUser"

	err := s.usersRepository.DeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
