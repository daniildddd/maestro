package hasher

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct {
	cfg Config
}

func NewBcryptHasher(cfg Config) (*BcryptHasher, error) {
	const maxPracticalCost = 14

	if cfg.Cost < bcrypt.MinCost || cfg.Cost > maxPracticalCost {
		return nil, fmt.Errorf(
			"hasher cost must be in [%d, %d], got %d",
			bcrypt.MinCost,
			maxPracticalCost,
			cfg.Cost,
		)
	}

	return &BcryptHasher{
		cfg: cfg,
	}, nil
}

func (h *BcryptHasher) Verify(hash string, plain string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		return fmt.Errorf("compare hash and password: %w", err)
	}

	return nil
}

func (h *BcryptHasher) Hash(password string) (string, error) {
	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(password), h.cfg.Cost)
	if err != nil {
		return "", fmt.Errorf("generate hash from password: %w", err)
	}

	return string(passwordHashed), nil
}
