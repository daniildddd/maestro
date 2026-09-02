package service_test

import (
	"context"

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
	ctx = reqctx.WithRequestID(ctx, uuid.New())
	ctx = reqctx.WithClientInfo(ctx, "203.0.113.7", "test-agent")

	return ctx
}
