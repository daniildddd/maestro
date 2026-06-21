package errs

import "errors"

const (
	errInvalidCredentials  = "invalid credentials"
	errInvalidRefreshToken = "invalid refresh token"
	errExpiredRefreshToken = "expired refresh token"
	errMissingRefreshToken = "missing refresh token"
	errInvalidInput        = "invalid input"
	errInvalidRequestBody  = "invalid request body"
)

var (
	ErrInvalidCredentials  = errors.New(errInvalidCredentials)
	ErrInvalidRefreshToken = errors.New(errInvalidRefreshToken)
	ErrExpiredRefreshToken = errors.New(errExpiredRefreshToken)
	ErrMissingRefreshToken = errors.New(errMissingRefreshToken)
	ErrInvalidInput        = errors.New(errInvalidInput)
	ErrInvalidRequestBody  = errors.New(errInvalidRequestBody)
)
