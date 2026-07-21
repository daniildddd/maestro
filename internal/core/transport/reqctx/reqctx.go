package reqctx

import "context"

type ctxKey struct {
	name string
}

var (
	keyRole = ctxKey{name: "role"}
)

func WithKeyRole(ctx context.Context, role string) context.Context {
	ctx = context.WithValue(ctx, keyRole, role)

	return ctx
}

func Role(ctx context.Context) string {
	v, ok := ctx.Value(keyRole).(string)
	if !ok {
		panic("role not found in context")
	}

	return v
}
