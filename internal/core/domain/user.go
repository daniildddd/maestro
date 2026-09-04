package domain

import (
	"errors"
	"fmt"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

var (
	ErrInvalidRole     = errors.New("role must be one of: user, admin")
	ErrInvalidPassword = errors.New("password must be 8-128 chars without control characters")
	ErrInvalidUsername = errors.New("username must be 3-32 chars from [a-zA-Z0-9_-]")
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type User struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

type UserFilter struct {
	Page     int
	Limit    int
	Username string
	UserRole string
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

	if err := u.validate(); err != nil {
		return User{}, fmt.Errorf("%s: %w", op, err)
	}

	return u, nil
}

func (u User) validate() error {
	const op = "domain.User.validate"

	if err := ValidateUsername(u.Username); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if u.Role != RoleUser && u.Role != RoleAdmin {
		return fmt.Errorf("%s: (username=%s): %w", op, u.Username, ErrInvalidRole)
	}

	return nil
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
	f.normalize()

	if err := f.validate(); err != nil {
		return nil, fmt.Errorf(
			"%s: %w", op, err,
		)
	}

	return f, nil
}

func (u *UserFilter) normalize() {
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

func (u *UserFilter) validate() error {
	const op = "domain.UserFilter.validate"

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
