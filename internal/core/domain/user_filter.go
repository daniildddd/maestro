package domain

import (
	"errors"
	"fmt"
	"regexp"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

var (
	ErrInvalidRole     = errors.New("role must be one of: user, admin")
	ErrInvalidUsername = errors.New("username must contain only letters, digits, '_' and '-'")
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type UserFilter struct {
	Page     int
	Limit    int
	Username string
	UserRole string
}

func NewUserFilter(
	page int,
	limit int,
	username string,
	userRole string,
) (UserFilter, error) {
	const op = "domain.NewUserFilter"

	f := UserFilter{
		Page:     page,
		Limit:    limit,
		Username: username,
		UserRole: userRole,
	}
	f.Normalize()

	if err := f.Validate(); err != nil {
		return UserFilter{}, fmt.Errorf(
			"%s: %w", op, err,
		)
	}

	return f, nil
}

func (u *UserFilter) Normalize() {
	if u.Page <= 0 {
		u.Page = 1
	}

	if u.Limit <= 0 {
		u.Limit = 20
	}

	if u.Limit > 100 {
		u.Limit = 100
	}
}

func (u *UserFilter) Validate() error {
	const op = "domain.UserFilter.Validate"

	if u.UserRole != "" && u.UserRole != RoleUser && u.UserRole != RoleAdmin {
		return fmt.Errorf("%s: %w", op, ErrInvalidRole)
	}

	if u.Username != "" && !usernameRegex.MatchString(u.Username) {
		return fmt.Errorf("%s: %w", op, ErrInvalidUsername)
	}

	return nil
}
