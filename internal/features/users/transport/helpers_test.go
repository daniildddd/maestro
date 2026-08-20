package transport_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/features/users/transport"
)

func nopLogger() *logger.Logger {
	return &logger.Logger{Logger: zap.NewNop()}
}

func newTestJSONRequest(t *testing.T, method, path, body, contentType string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return req.WithContext(logger.ToContext(req.Context(), nopLogger()))
}

func newUsersTestHandler(usersService transport.UsersService) *transport.UsersHTTPHandler {
	return transport.NewUsersHTTPHandler(usersService)
}

type errorResponseBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

//nolint:unparam // test helper mirrors domain.NewUser; any argument may vary per test
func mustNewUser(
	t *testing.T,
	id uuid.UUID,
	username string,
	passwordHash string,
	role string,
	createdAt time.Time,
	updatedAt *time.Time,
) domain.User {
	t.Helper()

	user, err := domain.NewUser(id, username, passwordHash, role, createdAt, updatedAt)
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	return user
}
