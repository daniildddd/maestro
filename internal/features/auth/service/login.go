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

func (s *AuthService) Login(
	ctx context.Context,
	username string,
	password string,
) (domain.TokenPair, error) {
	const op = "auth.service.Login"

	if err := domain.ValidateUsername(username); err != nil {
		return domain.TokenPair{}, fmt.Errorf(
			"%s: %w: %v",
			op,
			errs.ErrValidationFailed,
			err,
		)
	}

	user, err := s.authRepository.GetUserByName(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			s.recordLoginFailure(ctx, username)

			return domain.TokenPair{}, fmt.Errorf(
				"%s: user not found: %w: %v",
				op,
				errs.ErrInvalidCredentials,
				err,
			)
		}

		return domain.TokenPair{}, fmt.Errorf(
			"%s: get user: %w",
			op,
			err,
		)
	}

	err = s.passwordHasher.Verify(user.PasswordHash, password)
	if err != nil {
		s.recordLoginFailure(ctx, username)

		return domain.TokenPair{}, fmt.Errorf(
			"%s: verify password (username=%s): %w: %v",
			op,
			username,
			errs.ErrInvalidCredentials,
			err,
		)
	}

	accessToken, err := s.accessGen.Generate(user.ID, user.Role, user.Username)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf(
			"%s: generate access token: %w",
			op,
			err,
		)
	}

	rawToken, expiresAt, err := s.refreshGen.Generate()
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf(
			"%s: generate refresh token: %w",
			op,
			err,
		)
	}

	token := s.refreshGen.Hash(rawToken)

	refreshToken, err := domain.CreateRefreshToken(user.ID, token, expiresAt)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf(
			"%s: create refresh token: %w",
			op,
			err,
		)
	}

	err = s.authRepository.SaveRefreshToken(ctx, refreshToken)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf(
			"%s: save refresh token: %w",
			op,
			err,
		)
	}

	s.auditor.Record(ctx, domain.AuditEvent{
		ID:          uuid.New(),
		Action:      domain.ActionAuthLogin,
		Outcome:     domain.OutcomeSuccess,
		ActorID:     user.ID,
		ActorLogin:  user.Username,
		SubjectType: domain.AuditSubjectUser,
		SubjectID:   user.ID.String(),
		SubjectName: user.Username,
		RequestID:   reqctx.RequestID(ctx).String(),
		IP:          reqctx.ClientIP(ctx),
		UserAgent:   reqctx.UserAgent(ctx),
		CreatedAt:   time.Now(),
	})

	return domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawToken,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *AuthService) recordLoginFailure(
	ctx context.Context,
	username string,
) {
	s.auditor.Record(ctx, domain.AuditEvent{
		ID:            uuid.New(),
		Action:        domain.ActionAuthLogin,
		Outcome:       domain.OutcomeFailure,
		FailureReason: "invalid_credentials",
		SubjectType:   domain.AuditSubjectUser,
		SubjectName:   username,
		RequestID:     reqctx.RequestID(ctx).String(),
		IP:            reqctx.ClientIP(ctx),
		UserAgent:     reqctx.UserAgent(ctx),
		CreatedAt:     time.Now(),
	})
}
