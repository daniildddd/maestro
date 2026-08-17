package reqctx

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey struct {
	name string
}

var (
	keyRole      = ctxKey{name: "role"}
	keyUserID    = ctxKey{name: "user_id"}
	keyRequestID = ctxKey{name: "request_id"}
)

func WithRole(ctx context.Context, role string) context.Context {
	ctx = context.WithValue(ctx, keyRole, role)

	return ctx
}

func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	ctx = context.WithValue(ctx, keyUserID, userID)

	return ctx
}

func Role(ctx context.Context) string {
	v, ok := ctx.Value(keyRole).(string)
	if !ok {
		panic("role not found in context")
	}

	return v
}

func UserID(ctx context.Context) uuid.UUID {
	v, ok := ctx.Value(keyUserID).(uuid.UUID)
	if !ok {
		panic("userID not found in context")
	}

	return v
}

func WithRequestID(ctx context.Context, requestID uuid.UUID) context.Context {
	return context.WithValue(ctx, keyRequestID, requestID)
}

func RequestID(ctx context.Context) uuid.UUID {
	v, ok := ctx.Value(keyRequestID).(uuid.UUID)
	if !ok {
		panic("requestID not found in context")
	}

	return v
}
