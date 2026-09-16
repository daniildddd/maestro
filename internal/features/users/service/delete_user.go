package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
)

func (s *UsersService) DeleteUser(
	ctx context.Context,
	id uuid.UUID,
) error {
	const op = "users.service.DeleteUser"

	deleted, err := s.usersRepository.DeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	actorID := reqctx.UserID(ctx)

	s.auditor.Record(ctx, domain.AuditEvent{
		ID:          uuid.New(),
		Action:      domain.ActionUserDeleted,
		Outcome:     domain.OutcomeSuccess,
		ActorID:     actorID,
		ActorLogin:  reqctx.Username(ctx),
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
