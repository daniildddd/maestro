package reqctx

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey struct {
	name string
}

var (
	keyRole   = ctxKey{name: "role"}
	keyUserId = ctxKey{name: "user_id"}
)

func WithKeyRole(ctx context.Context, role string) context.Context {
	ctx = context.WithValue(ctx, keyRole, role)

	return ctx
}

func WithUserId(ctx context.Context, userId uuid.UUID) context.Context {
	ctx = context.WithValue(ctx, keyUserId, userId)

	return ctx
}

func Role(ctx context.Context) string {
	v, ok := ctx.Value(keyRole).(string)
	if !ok {
		panic("role not found in context")
	}

	return v
}

func UserId(ctx context.Context) uuid.UUID {
	v, ok := ctx.Value(keyUserId).(uuid.UUID)
	if !ok {
		panic("userId not found in context")
	}

	return v
}
