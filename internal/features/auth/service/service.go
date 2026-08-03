package service

import (
	"context"
	"time"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/google/uuid"
)

type AuthService struct {
	authRepository AuthRepository
	passwordHasher PasswordHasher
	accessGen      AccessTokenGenerator
	refreshGen     RefreshTokenManager
}

func NewAuthService(
	authRepository AuthRepository,
	passwordHasher PasswordHasher,
	accessGen AccessTokenGenerator,
	refreshGen RefreshTokenManager,
) *AuthService {
	return &AuthService{
		authRepository: authRepository,
		passwordHasher: passwordHasher,
		accessGen:      accessGen,
		refreshGen:     refreshGen,
	}
}

type PasswordHasher interface {
	Verify(
		hash string,
		plain string,
	) error
}

type AccessTokenGenerator interface {
	Generate(
		userID uuid.UUID,
		role string,
	) (string, error)
}

type RefreshTokenManager interface {
	Generate() (token string, expiresAt time.Time, err error)
	Hash(rawToken string) string
}

type AuthRepository interface {
	GetUser(
		ctx context.Context,
		username string,
	) (domain.User, error)

	SaveRefreshToken(
		ctx context.Context,
		token domain.RefreshToken,
	) error
}