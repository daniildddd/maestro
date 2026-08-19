package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/errs"
)

func (s *UsersService) DeleteMe(
	ctx context.Context,
	userID uuid.UUID,
	password string,
) error {
	const op = "users.service.DeleteMe"

	user, err := s.usersRepository.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("%s: get user: %w", op, err)
	}

	if err = s.passwordHasher.Verify(user.PasswordHash, password); err != nil {
		return fmt.Errorf(
			"%s: verify password (user_id=%s): %w: %v",
			op,
			userID,
			errs.ErrInvalidCredentials,
			err,
		)
	}

	if err = s.usersRepository.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
