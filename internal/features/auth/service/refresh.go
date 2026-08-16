package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
)

func (s *AuthService) Refresh(
	ctx context.Context,
	rawOldRefresh string,
) (domain.TokenPair, error) {
	const op = "auth.service.Refresh"

	oldHash := s.refreshGen.Hash(rawOldRefresh)

	storedToken, err := s.authRepository.GetRefreshTokenByHash(ctx, oldHash)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf(
			"%s: get refresh token: %w",
			op,
			err,
		)
	}

	if storedToken.IsExpired(time.Now()) {
		return domain.TokenPair{}, fmt.Errorf(
			"%s: user_id=%s: %w",
			op,
			storedToken.UserId,
			errs.ErrExpiredRefreshToken,
		)
	}

	user, err := s.authRepository.GetUserById(ctx, storedToken.UserId)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return domain.TokenPair{}, fmt.Errorf(
				"%s: user not found for refresh token(user_id=%s): %w",
				op,
				storedToken.UserId,
				errs.ErrInvalidRefreshToken,
			)
		}

		return domain.TokenPair{}, fmt.Errorf(
			"%s: get user by id(user_id=%s): %w",
			op,
			storedToken.UserId,
			err,
		)
	}

	err = s.authRepository.DeleteRefreshToken(ctx, storedToken.TokenHash)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf(
			"%s: delete refresh token: %w",
			op,
			err,
		)
	}

	rawNewRefresh, expiresAt, err := s.refreshGen.Generate()
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf(
			"%s: generate refresh token: %w",
			op,
			err,
		)
	}

	newHash := s.refreshGen.Hash(rawNewRefresh)

	newRefreshToken, err := domain.CreateRefreshToken(
		user.Id,
		newHash,
		expiresAt,
	)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf(
			"%s: create refresh token: %w",
			op,
			err,
		)
	}

	err = s.authRepository.SaveRefreshToken(ctx, newRefreshToken)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf(
			"%s: save refresh token: %w",
			op,
			err,
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

	return domain.NewTokenPair(
		accessToken,
		rawNewRefresh,
		expiresAt,
	), nil
}
