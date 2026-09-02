package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
)

func (s *UsersService) ChangeOwnPassword(
	ctx context.Context,
	userID uuid.UUID,
	oldPassword string,
	newPassword string,
) error {
	const op = "users.service.ChangeOwnPassword"

	if err := domain.ValidatePassword(newPassword); err != nil {
		return fmt.Errorf(
			"%s: (user_id=%s): %w: %v",
			op,
			userID,
			errs.ErrValidationFailed,
			err,
		)
	}

	user, err := s.usersRepository.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("%s: get user: %w", op, err)
	}

	if err = s.passwordHasher.Verify(user.PasswordHash, oldPassword); err != nil {
		return fmt.Errorf(
			"%s: verify old password (user_id=%s): %w: %v",
			op,
			userID,
			errs.ErrInvalidCredentials,
			err,
		)
	}

	passwordHash, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf(
			"%s: hash password (user_id=%s): %w",
			op,
			userID,
			err,
		)
	}

	if err = s.usersRepository.ChangePassword(ctx, userID, passwordHash); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	actorID := reqctx.UserID(ctx)

	s.auditor.Record(ctx, domain.AuditEvent{
		ID:          uuid.New(),
		Action:      domain.ActionUserPasswordChanged,
		Outcome:     domain.OutcomeSuccess,
		ActorID:     actorID,
		SubjectType: domain.AuditSubjectUser,
		SubjectID:   userID.String(),
		RequestID:   reqctx.RequestID(ctx).String(),
		IP:          reqctx.ClientIP(ctx),
		UserAgent:   reqctx.UserAgent(ctx),
		CreatedAt:   time.Now(),
	})

	return nil
}
