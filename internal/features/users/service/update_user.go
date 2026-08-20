package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
)

func (s *UsersService) UpdateUser(
	ctx context.Context,
	id uuid.UUID,
	username string,
) (domain.User, error) {
	const op = "users.service.UpdateUser"

	if err := domain.ValidateUsername(username); err != nil {
		return domain.User{}, fmt.Errorf(
			"%s: (user_id=%s): %w: %v",
			op,
			id,
			errs.ErrValidationFailed,
			err,
		)
	}

	user, err := s.usersRepository.UpdateUser(ctx, id, username)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
