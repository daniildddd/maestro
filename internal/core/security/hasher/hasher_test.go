package hasher_test

import (
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
		{"min cost 4 is valid", bcrypt.MinCost, false},
		{"max practical cost 14 is valid", 14, false},
		{"default cost 10 is valid", 10, false},
		{"cost below min fails", bcrypt.MinCost - 1, true},
		{"cost above max practical fails", 15, true},
		{"zero cost fails", 0, true},
		{"negative cost fails", -1, true},
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
		{"correct password verifies", validHash(t, h, "correct-password"), "correct-password", false},
		{"wrong password fails", validHash(t, h, "correct-password"), "wrong-password", true},
		{"malformed hash fails", "not-a-bcrypt-hash", "any-password", true},
		{"empty hash fails", "", "any-password", true},
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
	must.NotEqual(hash1, hash2, "bcrypt must produce different hashes for same password due to random salt")
}
