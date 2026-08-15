package hasher

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct {
	cfg Config
}

func NewBcryptHasher(cfg Config) (*BcryptHasher, error) {
	const op = "security.hasher.NewBcryptHasher"

	const maxPracticalCost = 14

	if cfg.Cost < bcrypt.MinCost || cfg.Cost > maxPracticalCost {
		return nil, fmt.Errorf(
			"%s: hasher cost must be in [%d, %d], got %d",
			op,
			bcrypt.MinCost,
			maxPracticalCost,
			cfg.Cost,
		)
	}

	return &BcryptHasher{
		cfg: cfg,
	}, nil
}

func (h *BcryptHasher) Verify(hash, plain string) error {
	const op = "security.hasher.Verify"

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		return fmt.Errorf("%s: compare hash and password: %w", op, err)
	}

	return nil
}

func (h *BcryptHasher) Hash(password string) (string, error) {
	const op = "security.hasher.Hash"

	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(password), h.cfg.Cost)
	if err != nil {
		return "", fmt.Errorf("%s: generate hash from password: %w", op, err)
	}

	return string(passwordHashed), nil
}
