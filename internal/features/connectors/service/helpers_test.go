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

type capturingAuditor struct {
	recorded []domain.AuditEvent
}

func (a *capturingAuditor) Record(_ context.Context, event domain.AuditEvent) {
	a.recorded = append(a.recorded, event)
}

var userID = uuid.New()

func testCtx() context.Context {
	ctx := context.Background()
	ctx = reqctx.WithUserID(ctx, userID)
	ctx = reqctx.WithRequestID(ctx, uuid.New())
	ctx = reqctx.WithClientInfo(ctx, "203.0.113.7", "test-agent")

	return ctx
}
