package domain

import "errors"

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

var ErrInvalidRole = errors.New("role must be one of: user, admin")
