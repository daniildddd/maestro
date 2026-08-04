package service

import (
	"context"
	"fmt"
)

func (s *AuthService) Logout(
	ctx context.Context,
	rawToken string,
) error {
	const op = "auth.service.Logout"

	token := s.refreshGen.Hash(rawToken)

	err := s.authRepository.DeleteRefreshToken(ctx, token)
	if err != nil {
		return fmt.Errorf(
			"%s: delete refresh token: %w",
			op,
			err,
		)
	}

	return nil
}
