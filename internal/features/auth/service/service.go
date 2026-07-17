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
	refreshGen     RefreshTokenGenerator
}

func NewAuthService(
	authRepository AuthRepository,
	passwordHasher PasswordHasher,
	accessGen AccessTokenGenerator,
	refreshGen RefreshTokenGenerator,
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

	Hash()
}

type AccessTokenGenerator interface {
	Generate(
		userID uuid.UUID,
		role string,
	) (string, error)
}

type RefreshTokenGenerator interface {
	Generate() (token string, expiresAt time.Time, err error)
}

type AuthRepository interface {
	GetUser(
		ctx context.Context,
		username string,
	) (domain.User, error)

	SaveRefreshToken(
		ctx context.Context,
		userID uuid.UUID,
		refreshToken string,
		expiresAt time.Time,
	) error
}
