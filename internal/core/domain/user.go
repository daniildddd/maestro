package domain

import (
	"errors"
	"fmt"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

func NewUser(
	id uuid.UUID,
	username string,
	passwordHash string,
	role string,
	createdAt time.Time,
	updatedAt *time.Time,
) (User, error) {
	const op = "core.domain.NewUser"

	u := User{
		ID:           id,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}

	if err := u.Validate(); err != nil {
		return User{}, fmt.Errorf("%s: %w", op, err)
	}

	return u, nil
}

func (u User) Validate() error {
	const op = "domain.User.Validate"

	if err := ValidateUsername(u.Username); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if u.Role != RoleUser && u.Role != RoleAdmin {
		return fmt.Errorf("%s: (username=%s): %w", op, u.Username, ErrInvalidRole)
	}

	return nil
}

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

var ErrInvalidRole = errors.New("role must be one of: user, admin")

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
) (*UserFilter, error) {
	const op = "core.domain.NewUserFilter"

	f := &UserFilter{
		Page:     page,
		Limit:    limit,
		Username: username,
		UserRole: userRole,
	}
	f.Normalize()

	if err := f.Validate(); err != nil {
		return nil, fmt.Errorf(
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

	if u.Username != "" {
		if err := ValidateUsername(u.Username); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}
