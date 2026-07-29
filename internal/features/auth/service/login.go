package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
)

func (s *AuthService) Login(
	ctx context.Context,
	username string,
	password string,
) (domain.TokenPair, error) {
	user, err := s.authRepository.GetUser(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return domain.TokenPair{}, fmt.Errorf(
				"get user from repository: %w",
				errs.ErrInvalidCredentials,
			)
		}
		return domain.TokenPair{}, fmt.Errorf(
			"get user from repository: %w", err)
	}

	err = s.passwordHasher.Verify(user.PasswordHash, password)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf(
			"verify password: %v: %w",
			err,
			errs.ErrInvalidCredentials,
		)
	}

	accessToken, err := s.accessGen.Generate(user.Id, user.Role)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf(
			"generate access token: %w", err)
	}

	rawToken, expiresAt, err := s.refreshGen.Generate()
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf(
			"generate refresh token: %w", err)
	}

	token := s.refreshGen.Hash(rawToken)

	refreshToken, err := domain.CreateRefreshToken(user.Id, token, expiresAt)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf(
			"create refresh token: %w", err,
		)
	}

	err = s.authRepository.SaveRefreshToken(ctx, refreshToken)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf(
			"save refresh token: %w", err)
	}

	return domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawToken,
		Username:     username,
		ExpiresAt:    expiresAt,
	}, nil
}
