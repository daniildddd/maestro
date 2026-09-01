package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
)

func (s *AuthService) Logout(
	ctx context.Context,
	rawToken string,
) error {
	const op = "auth.service.Logout"

	token := s.refreshGen.Hash(rawToken)

	storedToken, err := s.authRepository.GetRefreshTokenByHash(ctx, token)
	if err != nil {
		if errors.Is(err, errs.ErrRefreshTokenNotFound) {
			return nil
		}

		return fmt.Errorf(
			"%s: get refresh token: %w",
			op,
			err,
		)
	}

	err = s.authRepository.DeleteRefreshToken(ctx, token)
	if err != nil {
		return fmt.Errorf(
			"%s: delete refresh token: %w",
			op,
			err,
		)
	}

	s.auditor.Record(ctx, domain.AuditEvent{
		ID:          uuid.New(),
		Action:      domain.ActionAuthLogout,
		Outcome:     domain.OutcomeSuccess,
		ActorID:     storedToken.UserID,
		SubjectType: domain.AuditSubjectUser,
		SubjectID:   storedToken.UserID.String(),
		RequestID:   reqctx.RequestID(ctx).String(),
		IP:          reqctx.ClientIP(ctx),
		UserAgent:   reqctx.UserAgent(ctx),
		CreatedAt:   time.Now(),
	})

	return nil
}
