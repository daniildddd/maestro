package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	core_logger "github.com/daniildddd/maestro/internal/core/logger"
)

func newTestRequest(t *testing.T, method, path string, headers http.Header) *http.Request {
	t.Helper()

	r := httptest.NewRequest(method, path, http.NoBody)

	for k, values := range headers {
		for _, v := range values {
			r.Header.Add(k, v)
		}
	}

	return r
}

func nopLogger() *core_logger.Logger {
	return &core_logger.Logger{Logger: zap.NewNop()}
}

func newObservableLogger(t *testing.T, lvl zapcore.Level) (*core_logger.Logger, *observer.ObservedLogs) {
	t.Helper()

	core, recorder := observer.New(lvl)

	return &core_logger.Logger{Logger: zap.New(core)}, recorder
}
