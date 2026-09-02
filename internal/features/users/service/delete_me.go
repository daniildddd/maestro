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

	deleted, err := s.usersRepository.DeleteUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	actorID := reqctx.UserID(ctx)

	s.auditor.Record(ctx, domain.AuditEvent{
		ID:          uuid.New(),
		Action:      domain.ActionUserDeleted,
		Outcome:     domain.OutcomeSuccess,
		ActorID:     actorID,
		SubjectType: domain.AuditSubjectUser,
		SubjectID:   deleted.ID.String(),
		SubjectName: deleted.Username,
		StateBefore: map[string]any{
			stateKeyUsername: deleted.Username,
			stateKeyRole:     deleted.Role,
		},
		RequestID: reqctx.RequestID(ctx).String(),
		IP:        reqctx.ClientIP(ctx),
		UserAgent: reqctx.UserAgent(ctx),
		CreatedAt: time.Now(),
	})

	return nil
}
