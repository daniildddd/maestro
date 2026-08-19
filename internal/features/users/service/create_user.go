package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
)

func (s *UsersService) CreateUser(
	ctx context.Context,
	username string,
	password string,
	role string,
) (domain.User, error) {
	const op = "users.service.CreateUser"

	if err := domain.ValidatePassword(password); err != nil {
		return domain.User{}, fmt.Errorf(
			"%s: (username=%s): %w: %v",
			op,
			username,
			errs.ErrValidationFailed,
			err,
		)
	}

	passwordHash, err := s.passwordHasher.Hash(password)
	if err != nil {
		return domain.User{}, fmt.Errorf(
			"%s: hash password (username=%s): %w",
			op,
			username,
			err,
		)
	}

	user, err := domain.NewUser(
		uuid.New(),
		username,
		passwordHash,
		role,
		time.Now().UTC(),
		nil,
	)
	if err != nil {
		return domain.User{}, fmt.Errorf(
			"%s: build user: %w: %v",
			op,
			errs.ErrValidationFailed,
			err,
		)
	}

	created, err := s.usersRepository.CreateUser(ctx, user)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: create user: %w", op, err)
	}

	return created, nil
}
