package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
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
	GetUserByName(
		ctx context.Context,
		username string,
	) (domain.User, error)

	GetUserById(
		ctx context.Context,
		id uuid.UUID,
	) (domain.User, error)

	SaveRefreshToken(
		ctx context.Context,
		token domain.RefreshToken,
	) error

	GetRefreshTokenByHash(
		ctx context.Context,
		token string,
	) (domain.RefreshToken, error)

	DeleteRefreshToken(
		ctx context.Context,
		token string,
	) error
}
