package service_test

import (
	"context"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
	"github.com/daniildddd/maestro/internal/features/auth/service"
)

type noopAuditor struct{}

func (noopAuditor) Record(context.Context, domain.AuditEvent) {
}

func testCtx() context.Context {
	ctx := context.Background()
	ctx = reqctx.WithRequestID(ctx, uuid.New())
	ctx = reqctx.WithClientInfo(ctx, "203.0.113.7", "test-agent")

	return ctx
}

func newTestService(
	authRepository service.AuthRepository,
	passwordHasher service.PasswordHasher,
	accessGen service.AccessTokenGenerator,
	refreshGen service.RefreshTokenManager,
) *service.AuthService {
	return service.NewAuthService(
		authRepository,
		passwordHasher,
		accessGen,
		refreshGen,
		noopAuditor{},
	)
}
