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
	UserID uuid.UUID
	Role   string
}

func (m *Manager) Generate(
	userID uuid.UUID,
	role string,
) (string, error) {
	const op = "security.access.Generate"

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
		return "", fmt.Errorf("%s: sign token: %w", op, err)
	}

	return signed, nil
}

func (m *Manager) Verify(
	tokenStr string,
) (AuthUser, error) {
	const op = "security.access.Verify"

	var claims jwtClaims

	_, err := jwt.ParseWithClaims(
		tokenStr,
		&claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(
					"%s: unexpected method: %s",
					op,
					t.Header["alg"],
				)
			}

			return []byte(m.cfg.Secret), nil
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return AuthUser{}, fmt.Errorf("%s: token is expired: %w", op, errs.ErrAccessTokenExpired)
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return AuthUser{}, fmt.Errorf("%s: token signature: %w", op, errs.ErrAccessTokenInvalid)
		default:
			return AuthUser{}, fmt.Errorf("%s: verify token: %w", op, errs.ErrAccessTokenInvalid)
		}
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return AuthUser{}, fmt.Errorf("%s: invalid subject(uuid parse): %w", op, errs.ErrAccessTokenInvalid)
	}

	if userID == uuid.Nil {
		return AuthUser{}, fmt.Errorf("%s: subject is nil uuid: %w", op, errs.ErrAccessTokenInvalid)
	}

	if claims.Role == "" {
		return AuthUser{}, fmt.Errorf("%s: user role is missing: %w", op, errs.ErrAccessTokenInvalid)
	}

	return AuthUser{
		UserID: userID,
		Role:   claims.Role,
	}, nil
}
