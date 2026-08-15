package service_test

import (
	"github.com/daniildddd/maestro/internal/features/auth/service"
)

func newTestService(
	authRepository service.AuthRepository,
	passwordHasher service.PasswordHasher,
	accessGen service.AccessTokenGenerator,
	refreshGen service.RefreshTokenManager,
) *service.AuthService {
	return service.NewAuthService(
		authRepository,
		passwordHasher,
		accessGen,
		refreshGen,
	)
}
