package refresh_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/security/refresh"
)

func TestNewManager(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ttl     time.Duration
		wantErr bool
	}{
		{
			name:    "positive ttl is valid",
			ttl:     time.Hour,
			wantErr: false,
		},
		{
			name:    "small positive ttl is valid",
			ttl:     time.Second,
			wantErr: false,
		},
		{
			name:    "zero ttl fails",
			ttl:     0,
			wantErr: true,
		},
		{
			name:    "negative ttl fails",
			ttl:     -time.Second,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)

			m, err := refresh.NewManager(refresh.Config{TTL: tt.ttl})

			if tt.wantErr {
				must.Error(err)
				must.Nil(m)

				return
			}

			must.NoError(err)
			must.NotNil(m)
		})
	}
}

func TestManager_Generate(t *testing.T) {
	t.Parallel()
	must := require.New(t)

	m, err := refresh.NewManager(refresh.Config{TTL: time.Hour})
	must.NoError(err)

	before := time.Now()

	token, expiresAt, err := m.Generate()

	must.NoError(err)
	must.NotEmpty(token)
	must.True(expiresAt.After(before), "expiresAt must be after generation time")

	token2, _, err := m.Generate()
	must.NoError(err)
	must.NotEqual(token, token2, "generate must produce different tokens due to randomness")
}

func TestManager_Hash(t *testing.T) {
	t.Parallel()
	must := require.New(t)

	m, err := refresh.NewManager(refresh.Config{TTL: time.Hour})
	must.NoError(err)

	t.Run("hash matches SHA256", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		input := "test"
		expected := hashSHA256Hex(input)

		got := m.Hash(input)

		must.Equal(expected, got, "hash must be SHA256")
	})

	t.Run("empty string produces SHA256 of empty", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		got := m.Hash("")

		must.Equal(hashSHA256Hex(""), got, "hash of empty string must be SHA256 of empty")
		must.NotEmpty(got, "hash of empty string must not be empty")
	})

	t.Run("same input produces same hash", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		input := "test"

		h1 := m.Hash(input)
		h2 := m.Hash(input)

		must.Equal(h1, h2, "hash must be deterministic")
	})

	t.Run("different inputs produce different hashes", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		h1 := m.Hash("input1")
		h2 := m.Hash("input2")

		must.NotEqual(h1, h2, "different inputs must produce different hashes")
	})
}

func hashSHA256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))

	return hex.EncodeToString(sum[:])
}
