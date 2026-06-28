package errs

import "errors"

const (
	errInvalidCredentials  = "invalid credentials"
	errInvalidRefreshToken = "invalid refresh token"
	errExpiredRefreshToken = "expired refresh token"
	errMissingRefreshToken = "missing refresh token"
	errInvalidRequestBody  = "invalid request body"
	errUserNotFound        = "user not found"
)

var (
	ErrInvalidCredentials  = errors.New(errInvalidCredentials)
	ErrInvalidRefreshToken = errors.New(errInvalidRefreshToken)
	ErrExpiredRefreshToken = errors.New(errExpiredRefreshToken)
	ErrMissingRefreshToken = errors.New(errMissingRefreshToken)
	ErrInvalidRequestBody  = errors.New(errInvalidRequestBody)
	ErrUserNotFound        = errors.New(errUserNotFound)
)
