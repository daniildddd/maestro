package refresh

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

const tokenBytes = 32

type Manager struct {
	cfg Config
}

func NewManager(cfg Config) (*Manager, error) {
	if cfg.TTL <= 0 {
		return nil, fmt.Errorf(
			"refresh ttl must be greater than zero, got %d",
			cfg.TTL,
		)
	}

	return &Manager{
		cfg: cfg,
	}, nil
}

func (m *Manager) Generate() (token string, expiresAt time.Time, err error) {
	now := time.Now()

	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, fmt.Errorf(
			"read random bytes: %w",
			err,
		)
	}

	return base64.RawURLEncoding.EncodeToString(b), now.Add(m.cfg.TTL), nil
}
