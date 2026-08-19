package domain

import (
	"errors"
	"unicode/utf8"
)

var ErrInvalidPassword = errors.New("password must be 8-128 chars without control characters")

func ValidatePassword(password string) error {
	if n := utf8.RuneCountInString(password); n < 8 || n > 128 {
		return ErrInvalidPassword
	}

	for _, r := range password {
		if r < 0x20 || r == 0x7F {
			return ErrInvalidPassword
		}
	}

	return nil
}
