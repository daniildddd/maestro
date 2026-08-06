package access

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/errs"
)

type Manager struct {
	cfg Config
}

func NewManager(cfg Config) *Manager {
	return &Manager{
		cfg: cfg,
	}
}

type jwtClaims struct {
	jwt.RegisteredClaims

	Role string `json:"role"`
}

type AuthUser struct {
	UserId uuid.UUID
	Role   string
}

func (m *Manager) Generate(
	userID uuid.UUID,
	role string,
) (string, error) {
	now := time.Now()

	claims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.cfg.Issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.cfg.AccessTTL)),
		},
		Role: role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(m.cfg.Secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

func (m *Manager) Verify(
	tokenStr string,
) (AuthUser, error) {
	var claims jwtClaims

	_, err := jwt.ParseWithClaims(
		tokenStr,
		&claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(
					"unexpected method: %s", t.Header["alg"])
			}

			return []byte(m.cfg.Secret), nil
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return AuthUser{}, fmt.Errorf("token is expired: %w", errs.ErrAccessTokenExpired)
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return AuthUser{}, fmt.Errorf("token signature: %w", errs.ErrAccessTokenInvalid)
		default:
			return AuthUser{}, fmt.Errorf("verify token: %w", errs.ErrAccessTokenInvalid)
		}
	}

	userId, err := uuid.Parse(claims.Subject)
	if err != nil {
		return AuthUser{}, fmt.Errorf("invalid subject(uuid parse): %w", errs.ErrAccessTokenInvalid)
	}

	if userId == uuid.Nil {
		return AuthUser{}, fmt.Errorf("subject is nil uuid: %w", errs.ErrAccessTokenInvalid)
	}

	if claims.Role == "" {
		return AuthUser{}, fmt.Errorf("user role is missing: %w", errs.ErrAccessTokenInvalid)
	}

	return AuthUser{
		UserId: userId,
		Role:   claims.Role,
	}, nil
}
