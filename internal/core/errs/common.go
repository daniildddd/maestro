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
	Message:    "Session expired",
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
	Message:    "User with provided ID not found",
	LogLevel:   zapcore.InfoLevel,
}

var ErrUsernameConflict = &AppError{
	HTTPStatus: http.StatusConflict,
	Code:       "USERNAME_CONFLICT",
	Message:    "Username already taken",
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

var ErrInvalidContentType = &AppError{
	HTTPStatus: http.StatusBadRequest,
	Code:       "INVALID_CONTENT_TYPE",
	Message:    "Invalid content type",
	LogLevel:   zapcore.WarnLevel,
}

var ErrInvalidQueryParam = &AppError{
	HTTPStatus: http.StatusBadRequest,
	Code:       "INVALID_QUERY_PARAM",
	Message:    "Invalid query parameter",
	LogLevel:   zapcore.WarnLevel,
}

var ErrInvalidPathParam = &AppError{
	HTTPStatus: http.StatusBadRequest,
	Code:       "INVALID_PATH_PARAM",
	Message:    "Invalid path parameter",
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

var ErrKafkaConnectUnavailable = &AppError{
	HTTPStatus: http.StatusServiceUnavailable,
	Code:       "KAFKA_CONNECT_UNAVAILABLE",
	Message:    "Kafka Connect is unavailable",
	LogLevel:   zapcore.ErrorLevel,
}

var ErrConnectorNotFound = &AppError{
	HTTPStatus: http.StatusNotFound,
	Code:       "CONNECTOR_NOT_FOUND",
	Message:    "Connector not found",
	LogLevel:   zapcore.InfoLevel,
}

var ErrRebalanceInProgress = &AppError{
	HTTPStatus: http.StatusServiceUnavailable,
	Code:       "REBALANCE_IN_PROGRESS",
	Message:    "Kafka Connect is rebalancing, retry the request later",
	LogLevel:   zapcore.WarnLevel,
}

var ErrConnectorAlreadyExists = &AppError{
	HTTPStatus: http.StatusConflict,
	Code:       "CONNECTOR_ALREADY_EXISTS",
	Message:    "Connector already exists",
	LogLevel:   zapcore.InfoLevel,
}

var ErrConnectorTaskNotFound = &AppError{
	HTTPStatus: http.StatusNotFound,
	Code:       "NOT_FOUND",
	Message:    "Connector or task not found",
	LogLevel:   zapcore.InfoLevel,
}
