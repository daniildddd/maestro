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
	const op = "auth.service.Login"

	user, err := s.authRepository.GetUserByName(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return domain.TokenPair{}, fmt.Errorf(
				"%s: get user: %w: %w",
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
		return domain.TokenPair{}, fmt.Errorf(
			"%s: verify password: %w: %w",
			op,
			err,
			errs.ErrInvalidCredentials,
		)
	}

	accessToken, err := s.accessGen.Generate(user.Id, user.Role)
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

	refreshToken, err := domain.CreateRefreshToken(user.Id, token, expiresAt)
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

	return domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawToken,
		ExpiresAt:    expiresAt,
	}, nil
}
