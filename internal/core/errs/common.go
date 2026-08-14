package errs

import (
	"errors"
	"net/http"

	"go.uber.org/zap/zapcore"
)

const codeInvalidRefreshToken = "INVALID_REFRESH_TOKEN"

type AppError struct {
	HTTPStatus int
	Code       string
	Message    string
	LogLevel   zapcore.Level
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Is(target error) bool {
	var appErr *AppError

	if errors.As(target, &appErr) {
		return appErr.Code == e.Code
	}

	return false
}

var ErrInvalidRefreshToken = &AppError{
	HTTPStatus: http.StatusUnauthorized,
	Code:       codeInvalidRefreshToken,
	Message:    "Invalid refresh token",
	LogLevel:   zapcore.WarnLevel,
}

var ErrExpiredRefreshToken = &AppError{
	HTTPStatus: http.StatusUnauthorized,
	Code:       codeInvalidRefreshToken,
	Message:    "Expired refresh token",
	LogLevel:   zapcore.InfoLevel,
}

var ErrMissingRefreshToken = &AppError{
	HTTPStatus: http.StatusUnauthorized,
	Code:       codeInvalidRefreshToken,
	Message:    "Refresh token is required",
	LogLevel:   zapcore.WarnLevel,
}

var ErrAccessTokenExpired = &AppError{
	HTTPStatus: http.StatusUnauthorized,
	Code:       "ACCESS_TOKEN_EXPIRED",
	Message:    "Access token expired",
	LogLevel:   zapcore.InfoLevel,
}

var ErrAccessTokenInvalid = &AppError{
	HTTPStatus: http.StatusUnauthorized,
	Code:       "ACCESS_TOKEN_INVALID",
	Message:    "Invalid access token",
	LogLevel:   zapcore.WarnLevel,
}

var ErrAccessTokenMissing = &AppError{
	HTTPStatus: http.StatusUnauthorized,
	Code:       "ACCESS_TOKEN_MISSING",
	Message:    "Access token missing",
	LogLevel:   zapcore.WarnLevel,
}

var ErrUserNotFound = &AppError{
	HTTPStatus: http.StatusNotFound,
	Code:       "USER_NOT_FOUND",
	Message:    "User not found",
	LogLevel:   zapcore.InfoLevel,
}

var ErrForbidden = &AppError{
	HTTPStatus: http.StatusForbidden,
	Code:       "FORBIDDEN",
	Message:    "Forbidden",
	LogLevel:   zapcore.WarnLevel,
}

var ErrRefreshTokenNotFound = &AppError{
	HTTPStatus: http.StatusNotFound,
	Code:       "REFRESH_TOKEN_NOT_FOUND",
	Message:    "Refresh token not found",
	LogLevel:   zapcore.WarnLevel,
}

var ErrInvalidCredentials = &AppError{
	HTTPStatus: http.StatusBadRequest,
	Code:       "INVALID_CREDENTIALS",
	Message:    "Invalid credentials",
	LogLevel:   zapcore.InfoLevel,
}

var ErrInvalidRequestBody = &AppError{
	HTTPStatus: http.StatusBadRequest,
	Code:       "INVALID_REQUEST_BODY",
	Message:    "Invalid request body",
	LogLevel:   zapcore.WarnLevel,
}

var ErrValidationFailed = &AppError{
	HTTPStatus: http.StatusBadRequest,
	Code:       "VALIDATION_FAILED",
	Message:    "Validation failed",
	LogLevel:   zapcore.InfoLevel,
}

var ErrInternal = &AppError{
	HTTPStatus: http.StatusInternalServerError,
	Code:       "INTERNAL_ERROR",
	Message:    "Internal server error",
	LogLevel:   zapcore.ErrorLevel,
}
