package refresh

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"
)

const tokenBytes = 32

type Manager struct {
	cfg Config
}

func NewManager(cfg Config) (*Manager, error) {
	const op = "security.refresh.NewManager"

	if cfg.TTL <= 0 {
		return nil, fmt.Errorf(
			"%s: refresh ttl must be greater than zero, got %d",
			op,
			cfg.TTL,
		)
	}

	return &Manager{
		cfg: cfg,
	}, nil
}

func (m *Manager) Generate() (string, time.Time, error) {
	const op = "security.refresh.Generate"

	now := time.Now()

	var b [tokenBytes]byte

	if _, err := rand.Read(b[:]); err != nil {
		return "", time.Time{}, fmt.Errorf(
			"%s: read random bytes: %w",
			op,
			err,
		)
	}

	return base64.RawURLEncoding.EncodeToString(b[:]), now.Add(m.cfg.TTL), nil
}

func (m *Manager) Hash(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))

	return hex.EncodeToString(sum[:])
}
