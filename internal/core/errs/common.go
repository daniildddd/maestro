package errs

import "errors"

const (
	errInvalidCredentials  = "invalid credentials"
	errInvalidRefreshToken = "refresh token invalid"
	errExpiredRefreshToken = "refresh token expired"
	errMissingRefreshToken = "refresh token missing"
)

var (
	ErrInvalidCredentials  = errors.New(errInvalidCredentials)
	ErrInvalidRefreshToken = errors.New(errInvalidRefreshToken)
	ErrExpiredRefreshToken = errors.New(errExpiredRefreshToken)
	ErrMissingRefreshToken = errors.New(errMissingRefreshToken)
)
