package hasher_test

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/security/hasher"
)

func TestNewBcryptHasher(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cost    int
		wantErr bool
	}{
		{
			name:    "min cost 4 is valid",
			cost:    bcrypt.MinCost,
			wantErr: false,
		},
		{
			name:    "max practical cost 14 is valid",
			cost:    14,
			wantErr: false,
		},
		{
			name:    "default cost 10 is valid",
			cost:    10,
			wantErr: false,
		},
		{
			name:    "cost below min fails",
			cost:    bcrypt.MinCost - 1,
			wantErr: true,
		},
		{
			name:    "cost above max practical fails",
			cost:    15,
			wantErr: true,
		},
		{
			name:    "zero cost fails",
			cost:    0,
			wantErr: true,
		},
		{
			name:    "negative cost fails",
			cost:    -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)

			h, err := hasher.NewBcryptHasher(hasher.Config{Cost: tt.cost})

			if tt.wantErr {
				must.Error(err)
				must.Nil(h)

				return
			}

			must.NoError(err)
			must.NotNil(h)
		})
	}
}

func validHash(t *testing.T, h *hasher.BcryptHasher, password string) string {
	t.Helper()
	must := require.New(t)

	hash, err := h.Hash(password)
	must.NoError(err)

	return hash
}

func TestBcryptHasher_Verify(t *testing.T) {
	t.Parallel()
	must := require.New(t)

	h, err := hasher.NewBcryptHasher(hasher.Config{Cost: bcrypt.MinCost})
	must.NoError(err)

	tests := []struct {
		name    string
		hash    string
		plain   string
		wantErr bool
	}{
		{
			name:    "correct password verifies",
			hash:    validHash(t, h, "correct-password"),
			plain:   "correct-password",
			wantErr: false,
		},
		{
			name:    "wrong password fails",
			hash:    validHash(t, h, "correct-password"),
			plain:   "wrong-password",
			wantErr: true,
		},
		{
			name:    "malformed hash fails",
			hash:    "not-a-bcrypt-hash",
			plain:   "any-password",
			wantErr: true,
		},
		{
			name:    "empty hash fails",
			hash:    "",
			plain:   "any-password",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)

			err := h.Verify(tt.hash, tt.plain)

			if tt.wantErr {
				must.Error(err)

				return
			}

			must.NoError(err)
		})
	}
}

func TestBcryptHasher_Hash(t *testing.T) {
	t.Parallel()
	must := require.New(t)

	h, err := hasher.NewBcryptHasher(hasher.Config{Cost: bcrypt.MinCost})
	must.NoError(err)

	hash1, err := h.Hash("my-password")

	must.NoError(err)
	must.NotEmpty(hash1)
	must.NotEqual("my-password", hash1)

	hash2, err := h.Hash("my-password")

	must.NoError(err)
	must.NotEqual(
		hash1,
		hash2,
		"bcrypt must produce different hashes for same password due to random salt",
	)
}

func TestBcryptHasher_Hash_PasswordLength(t *testing.T) {
	t.Parallel()
	must := require.New(t)

	h, err := hasher.NewBcryptHasher(hasher.Config{Cost: bcrypt.MinCost})
	must.NoError(err)

	tests := []struct {
		name      string
		password  string
		wantErrIs error
	}{
		{
			name:     "72 bytes password hashes successfully",
			password: strings.Repeat("a", 72),
		},
		{
			name:      "73 bytes password fails",
			password:  strings.Repeat("a", 73),
			wantErrIs: bcrypt.ErrPasswordTooLong,
		},
		{
			name:      "long password fails",
			password:  strings.Repeat("password-", 10),
			wantErrIs: bcrypt.ErrPasswordTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)

			hash, err := h.Hash(tt.password)

			if tt.wantErrIs != nil {
				must.Error(err)
				must.ErrorIs(err, tt.wantErrIs)
				must.Empty(hash)

				return
			}

			must.NoError(err)
			must.NotEmpty(hash)
		})
	}
}
