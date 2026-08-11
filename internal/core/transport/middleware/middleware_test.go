package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/transport/middleware"
)

type testHandler struct{ body string }

func (h *testHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte(h.body)) //nolint:errcheck // test-only: httptest ResponseWriter. Write never fails
}

func makeTracingMiddleware(id string) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(id + "->")) //nolint:errcheck // test-only: httptest ResponseWriter. Write never fails
			next.ServeHTTP(w, r)
		})
	}
}

func TestChainMiddleware(t *testing.T) {
	t.Parallel()

	finalHandler := &testHandler{body: "done"}

	t.Run("no middleware returns handler as-is", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		chained := middleware.ChainMiddleware(finalHandler)

		must.Same(finalHandler, chained, "empty chain must return same handler without wrapping")

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rec := httptest.NewRecorder()
		chained.ServeHTTP(rec, req)
		must.Equal("done", rec.Body.String())
	})

	t.Run("single middleware wraps handler", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		chained := middleware.ChainMiddleware(finalHandler, makeTracingMiddleware("MW1"))

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rec := httptest.NewRecorder()
		chained.ServeHTTP(rec, req)
		must.Equal("MW1->done", rec.Body.String())
	})

	t.Run("middleware applied in order", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		chained := middleware.ChainMiddleware(
			finalHandler,
			makeTracingMiddleware("MW1"),
			makeTracingMiddleware("MW2"),
			makeTracingMiddleware("MW3"),
		)

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rec := httptest.NewRecorder()
		chained.ServeHTTP(rec, req)
		must.Equal("MW1->MW2->MW3->done", rec.Body.String())
	})
}
