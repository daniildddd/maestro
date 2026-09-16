package reqctx_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
)

type testCtxKey int

const wrongKey testCtxKey = 1

func TestRole(t *testing.T) {
	t.Parallel()

	t.Run("returns role when set in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ctx := reqctx.WithRole(context.Background(), "admin")

		is.Equal("admin", reqctx.Role(ctx))
	})

	t.Run("overwrites role when set multiple times", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ctx := reqctx.WithRole(context.Background(), "user")
		ctx = reqctx.WithRole(ctx, "admin")

		is.Equal("admin", reqctx.Role(ctx))
	})

	t.Run("panics when role not in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ctx := context.Background()

		is.PanicsWithValue("role not found in context", func() {
			reqctx.Role(ctx)
		})
	})

	t.Run("panics when wrong type in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ctx := context.WithValue(context.Background(), wrongKey, 123)

		is.PanicsWithValue("role not found in context", func() {
			reqctx.Role(ctx)
		})
	})
}

func TestUserId(t *testing.T) {
	t.Parallel()

	t.Run("returns userID when set in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		userID := uuid.New()
		ctx := reqctx.WithUserID(context.Background(), userID)

		is.Equal(userID, reqctx.UserID(ctx))
	})

	t.Run("overwrites userID when set multiple times", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		firstId := uuid.New()
		secondID := uuid.New()
		ctx := reqctx.WithUserID(context.Background(), firstId)
		ctx = reqctx.WithUserID(ctx, secondID)

		is.Equal(secondID, reqctx.UserID(ctx))
	})

	t.Run("panics when userID not in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ctx := context.Background()

		is.PanicsWithValue("userID not found in context", func() {
			reqctx.UserID(ctx)
		})
	})

	t.Run("panics when wrong type in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ctx := context.WithValue(context.Background(), wrongKey, "not-a-uuid")

		is.PanicsWithValue("userID not found in context", func() {
			reqctx.UserID(ctx)
		})
	})
}

func TestUsername(t *testing.T) {
	t.Parallel()

	t.Run("returns username when set in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ctx := reqctx.WithUsername(context.Background(), "alice")

		is.Equal("alice", reqctx.Username(ctx))
	})

	t.Run("overwrites username when set multiple times", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ctx := reqctx.WithUsername(context.Background(), "alice")
		ctx = reqctx.WithUsername(ctx, "bob")

		is.Equal("bob", reqctx.Username(ctx))
	})

	t.Run("panics when username not in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ctx := context.Background()

		is.PanicsWithValue("username not found in context", func() {
			reqctx.Username(ctx)
		})
	})
}

func TestRequestId(t *testing.T) {
	t.Parallel()

	t.Run("returns requestID when set in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		requestID := uuid.New()
		ctx := reqctx.WithRequestID(context.Background(), requestID)

		is.Equal(requestID, reqctx.RequestID(ctx))
	})

	t.Run("overwrites requestID when set multiple times", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		firstId := uuid.New()
		secondID := uuid.New()
		ctx := reqctx.WithRequestID(context.Background(), firstId)
		ctx = reqctx.WithRequestID(ctx, secondID)

		is.Equal(secondID, reqctx.RequestID(ctx))
	})

	t.Run("panics when requestID not in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ctx := context.Background()

		is.PanicsWithValue("requestID not found in context", func() {
			reqctx.RequestID(ctx)
		})
	})

	t.Run("panics when wrong type in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ctx := context.WithValue(context.Background(), wrongKey, 123)

		is.PanicsWithValue("requestID not found in context", func() {
			reqctx.RequestID(ctx)
		})
	})
}

func TestClientInfo(t *testing.T) {
	t.Parallel()

	t.Run("returns ip and user agent when set in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ctx := reqctx.WithClientInfo(context.Background(), "192.168.1.10", "curl/8.7.1")

		is.Equal("192.168.1.10", reqctx.ClientIP(ctx))
		is.Equal("curl/8.7.1", reqctx.UserAgent(ctx))
	})

	t.Run("overwrites client info when set multiple times", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ctx := reqctx.WithClientInfo(context.Background(), "10.0.0.1", "first")
		ctx = reqctx.WithClientInfo(ctx, "10.0.0.2", "second")

		is.Equal("10.0.0.2", reqctx.ClientIP(ctx))
		is.Equal("second", reqctx.UserAgent(ctx))
	})

	t.Run("accepts empty values without panic", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ctx := reqctx.WithClientInfo(context.Background(), "", "")

		is.Empty(reqctx.ClientIP(ctx))
		is.Empty(reqctx.UserAgent(ctx))
	})

	t.Run("panics when client ip not in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		is.PanicsWithValue("clientIP not found in context", func() {
			reqctx.ClientIP(context.Background())
		})
	})

	t.Run("panics when user agent not in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		is.PanicsWithValue("userAgent not found in context", func() {
			reqctx.UserAgent(context.Background())
		})
	})
}
