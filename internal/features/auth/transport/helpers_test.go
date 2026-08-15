package transport_test

import (
	"go.uber.org/zap"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/features/auth/transport"
)

func nopLogger() *logger.Logger {
	return &logger.Logger{Logger: zap.NewNop()}
}

func newTestHandler(authService transport.AuthService, cfg transport.Config) *transport.AuthHTTPHandler {
	return transport.NewAuthHTTPHandler(authService, cfg)
}

type errorResponseBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type refreshResponseBody struct {
	AccessToken string `json:"access_token"`
}
