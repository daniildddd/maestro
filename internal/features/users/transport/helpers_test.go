package transport_test

import (
	"go.uber.org/zap"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/features/users/transport"
)

func nopLogger() *logger.Logger {
	return &logger.Logger{Logger: zap.NewNop()}
}

func newUsersTestHandler(usersService transport.UsersService) *transport.UsersHTTPHandler {
	return transport.NewUsersHTTPHandler(usersService)
}

type errorResponseBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
