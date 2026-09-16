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

	before, after, err := s.usersRepository.UpdateUser(ctx, id, username)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	actorID := reqctx.UserID(ctx)

	s.auditor.Record(ctx, domain.AuditEvent{
		ID:          uuid.New(),
		Action:      domain.ActionUserUpdated,
		Outcome:     domain.OutcomeSuccess,
		ActorID:     actorID,
		ActorLogin:  reqctx.Username(ctx),
		SubjectType: domain.AuditSubjectUser,
		SubjectID:   after.ID.String(),
		SubjectName: after.Username,
		StateBefore: map[string]any{
			stateKeyUsername: before.Username,
			stateKeyRole:     before.Role,
		},
		StateAfter: map[string]any{
			stateKeyUsername: after.Username,
			stateKeyRole:     after.Role,
		},
		RequestID: reqctx.RequestID(ctx).String(),
		IP:        reqctx.ClientIP(ctx),
		UserAgent: reqctx.UserAgent(ctx),
		CreatedAt: time.Now(),
	})

	return after, nil
}
