package domain

import (
	"errors"
	"fmt"
	"regexp"
)

var ErrInvalidUsername = errors.New("username must be 3-32 chars from [a-zA-Z0-9_-]")

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func ValidateUsername(username string) error {
	const op = "core.domain.ValidateUsername"

	if len(username) < 3 || len(username) > 32 {
		return fmt.Errorf("%s: username=%s: %w", op, username, ErrInvalidUsername)
	}

	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("%s: username=%s: %w", op, username, ErrInvalidUsername)
	}

	return nil
}
