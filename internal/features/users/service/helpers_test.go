package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
)

type noopAuditor struct{}

func (noopAuditor) Record(context.Context, domain.AuditEvent) {
}

var userID = uuid.New()

func testCtx() context.Context {
	ctx := context.Background()
	ctx = reqctx.WithUserID(ctx, userID)
	ctx = reqctx.WithUsername(ctx, "alice")
	ctx = reqctx.WithRequestID(ctx, uuid.New())
	ctx = reqctx.WithClientInfo(ctx, "203.0.113.7", "test-agent")

	return ctx
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
