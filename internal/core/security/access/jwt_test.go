package access_test

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/security/access"
)

type testClaims struct {
	jwt.RegisteredClaims

	Role string `json:"role"`
}

func signToken(
	t *testing.T,
	cfg access.Config,
	subject string,
	role string,
	expiresAt time.Time,
) string {
	t.Helper()
	must := require.New(t)

	claims := testClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		Role: role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.Secret))
	must.NoError(err)

	return signed
}

func expiredToken(t *testing.T, cfg access.Config) string {
	t.Helper()

	return signToken(t, cfg, uuid.New().String(), "admin", time.Now().Add(-time.Minute))
}

func wrongSecretToken(t *testing.T, cfg access.Config, userID uuid.UUID, role string) string {
	t.Helper()

	other := cfg
	other.Secret = "different-secret"

	return signToken(t, other, userID.String(), role, time.Now().Add(time.Minute))
}

func wrongAlgToken(t *testing.T, userID uuid.UUID, role string) string {
	t.Helper()
	must := require.New(t)

	claims := testClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		Role: role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType) //nolint:gosec // test-only: verify must reject non-HMAC
	must.NoError(err)

	return signed
}

func validToken(t *testing.T, m *access.Manager, userID uuid.UUID, role string) string {
	t.Helper()
	must := require.New(t)

	token, err := m.Generate(userID, role)
	must.NoError(err)

	return token
}

func TestManager_Generate(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	cfg := access.Config{
		Secret:    "test-secret",
		AccessTTL: time.Minute,
		Issuer:    "maestro",
	}
	manager := access.NewManager(cfg)

	userID := uuid.New()
	role := "admin"

	token, err := manager.Generate(userID, role)

	must.NoError(err)
	is.NotEmpty(token)
	is.Len(strings.Split(token, "."), 3)

	var claims testClaims

	parsed, err := jwt.ParseWithClaims(token, &claims, func(_ *jwt.Token) (any, error) {
		return []byte(cfg.Secret), nil
	})
	must.NoError(err)
	is.True(parsed.Valid)
	is.Equal(cfg.Issuer, claims.Issuer)
	is.Equal(userID.String(), claims.Subject)
	is.Equal(role, claims.Role)
	is.NotNil(claims.ExpiresAt)
	is.NotNil(claims.IssuedAt)
}

func TestManager_Verify(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	cfg := access.Config{
		Secret:    "test-secret",
		AccessTTL: time.Minute,
		Issuer:    "maestro",
	}
	manager := access.NewManager(cfg)

	validUserID := uuid.New()
	validRole := "admin"

	tests := []struct {
		name     string
		token    string
		target   error
		wantUser *access.AuthUser
	}{
		{
			name:     "success returns auth user",
			token:    validToken(t, manager, validUserID, validRole),
			wantUser: &access.AuthUser{UserId: validUserID, Role: validRole},
		},
		{
			name:   "expired token maps to ErrAccessTokenExpired",
			token:  expiredToken(t, cfg),
			target: errs.ErrAccessTokenExpired,
		},
		{
			name:   "invalid signature maps to ErrAccessTokenInvalid",
			token:  wrongSecretToken(t, cfg, validUserID, validRole),
			target: errs.ErrAccessTokenInvalid,
		},
		{
			name:   "wrong signing algorithm maps to ErrAccessTokenInvalid",
			token:  wrongAlgToken(t, validUserID, validRole),
			target: errs.ErrAccessTokenInvalid,
		},
		{
			name:   "malformed token maps to ErrAccessTokenInvalid",
			token:  "not-a-jwt",
			target: errs.ErrAccessTokenInvalid,
		},
		{
			name:   "invalid subject (not uuid) maps to ErrAccessTokenInvalid",
			token:  signToken(t, cfg, "not-a-uuid", validRole, time.Now().Add(time.Minute)),
			target: errs.ErrAccessTokenInvalid,
		},
		{
			name:   "nil uuid subject maps to ErrAccessTokenInvalid",
			token:  signToken(t, cfg, uuid.Nil.String(), validRole, time.Now().Add(time.Minute)),
			target: errs.ErrAccessTokenInvalid,
		},
		{
			name:   "empty role maps to ErrAccessTokenInvalid",
			token:  signToken(t, cfg, validUserID.String(), "", time.Now().Add(time.Minute)),
			target: errs.ErrAccessTokenInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user, err := manager.Verify(tt.token)

			if tt.target != nil {
				must.Error(err)
				must.ErrorIs(err, tt.target)
				is.Equal(access.AuthUser{}, user)

				return
			}

			must.NoError(err)
			is.Equal(*tt.wantUser, user)
		})
	}
}
