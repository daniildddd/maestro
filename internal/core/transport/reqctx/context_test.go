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

	t.Run("returns userId when set in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		userId := uuid.New()
		ctx := reqctx.WithUserId(context.Background(), userId)

		is.Equal(userId, reqctx.UserId(ctx))
	})

	t.Run("overwrites userId when set multiple times", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		firstId := uuid.New()
		secondId := uuid.New()
		ctx := reqctx.WithUserId(context.Background(), firstId)
		ctx = reqctx.WithUserId(ctx, secondId)

		is.Equal(secondId, reqctx.UserId(ctx))
	})

	t.Run("panics when userId not in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ctx := context.Background()

		is.PanicsWithValue("userId not found in context", func() {
			reqctx.UserId(ctx)
		})
	})

	t.Run("panics when wrong type in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ctx := context.WithValue(context.Background(), wrongKey, "not-a-uuid")

		is.PanicsWithValue("userId not found in context", func() {
			reqctx.UserId(ctx)
		})
	})
}

func TestRequestId(t *testing.T) {
	t.Parallel()

	t.Run("returns requestId when set in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		requestId := uuid.New()
		ctx := reqctx.WithRequestId(context.Background(), requestId)

		is.Equal(requestId, reqctx.RequestId(ctx))
	})

	t.Run("overwrites requestId when set multiple times", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		firstId := uuid.New()
		secondId := uuid.New()
		ctx := reqctx.WithRequestId(context.Background(), firstId)
		ctx = reqctx.WithRequestId(ctx, secondId)

		is.Equal(secondId, reqctx.RequestId(ctx))
	})

	t.Run("panics when requestId not in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ctx := context.Background()

		is.PanicsWithValue("requestId not found in context", func() {
			reqctx.RequestId(ctx)
		})
	})

	t.Run("panics when wrong type in context", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ctx := context.WithValue(context.Background(), wrongKey, 123)

		is.PanicsWithValue("requestId not found in context", func() {
			reqctx.RequestId(ctx)
		})
	})
}
