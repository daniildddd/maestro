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

	actorID := reqctx.UserID(ctx)

	s.auditor.Record(ctx, domain.AuditEvent{
		ID:          uuid.New(),
		Action:      domain.ActionUserPasswordChanged,
		Outcome:     domain.OutcomeSuccess,
		ActorID:     actorID,
		ActorLogin:  reqctx.Username(ctx),
		SubjectType: domain.AuditSubjectUser,
		SubjectID:   id.String(),
		RequestID:   reqctx.RequestID(ctx).String(),
		IP:          reqctx.ClientIP(ctx),
		UserAgent:   reqctx.UserAgent(ctx),
		CreatedAt:   time.Now(),
	})

	return nil
}
