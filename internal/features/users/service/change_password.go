package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
)

func (s *UsersService) ChangePassword(
	ctx context.Context,
	id uuid.UUID,
	newPassword string,
) error {
	const op = "users.service.ChangePassword"

	if err := domain.ValidatePassword(newPassword); err != nil {
		return fmt.Errorf(
			"%s: (user_id=%s): %w: %v",
			op,
			id,
			errs.ErrValidationFailed,
			err,
		)
	}

	passwordHash, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf(
			"%s: hash password (user_id=%s): %w",
			op,
			id,
			err,
		)
	}

	if err = s.usersRepository.ChangePassword(ctx, id, passwordHash); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
