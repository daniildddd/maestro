package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/require"

	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/middleware"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func makeTracingMiddleware(id string) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(id + "->")) //nolint:errcheck // test-only: httptest ResponseWriter.Write never fails
			next.ServeHTTP(w, r)
		})
	}
}

func nopLogger() *core_logger.Logger {
	return &core_logger.Logger{Logger: zap.NewNop()}
}

func TestRoute_WithMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("returns handler as-is when no middleware", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("done")) //nolint:errcheck // test-only: httptest ResponseWriter.Write never fails
		})

		route := &Route{
			Method:  http.MethodGet,
			Path:    "/",
			Handler: handler,
		}

		chained := route.WithMiddleware()

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rec := httptest.NewRecorder()
		chained.ServeHTTP(rec, req)

		must.Equal("done", rec.Body.String())
	})

	t.Run("applies single middleware", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("done")) //nolint:errcheck // test-only: httptest ResponseWriter.Write never fails
		})

		route := &Route{
			Method:     http.MethodGet,
			Path:       "/",
			Handler:    handler,
			Middleware: []middleware.Middleware{makeTracingMiddleware("MW1")},
		}

		chained := route.WithMiddleware()

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rec := httptest.NewRecorder()
		chained.ServeHTTP(rec, req)

		must.Equal("MW1->done", rec.Body.String())
	})

	t.Run("applies middleware in order", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("done")) //nolint:errcheck // test-only: httptest ResponseWriter.Write never fails
		})

		route := &Route{
			Method:  http.MethodGet,
			Path:    "/",
			Handler: handler,
			Middleware: []middleware.Middleware{
				makeTracingMiddleware("MW1"),
				makeTracingMiddleware("MW2"),
				makeTracingMiddleware("MW3"),
			},
		}

		chained := route.WithMiddleware()

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rec := httptest.NewRecorder()
		chained.ServeHTTP(rec, req)

		must.Equal("MW1->MW2->MW3->done", rec.Body.String())
	})

	t.Run("applies RequireRole when Roles set", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		route := &Route{
			Method:  http.MethodGet,
			Path:    "/",
			Handler: handler,
			Roles:   []string{"admin"},
		}

		chained := route.WithMiddleware()

		t.Run("allowed role passes through", func(t *testing.T) {
			t.Parallel()
			must := require.New(t)

			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			ctx := core_logger.ToContext(req.Context(), nopLogger())
			ctx = reqctx.WithRole(ctx, "admin")
			req = req.WithContext(ctx)

			rec := httptest.NewRecorder()
			rw := core_http_response.NewResponseWriter(rec)

			chained.ServeHTTP(rw, req)

			must.Equal(http.StatusOK, rec.Code)
		})

		t.Run("forbidden role rejected", func(t *testing.T) {
			t.Parallel()
			must := require.New(t)

			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			ctx := core_logger.ToContext(req.Context(), nopLogger())
			ctx = reqctx.WithRole(ctx, "user")
			req = req.WithContext(ctx)

			rec := httptest.NewRecorder()
			rw := core_http_response.NewResponseWriter(rec)

			chained.ServeHTTP(rw, req)

			must.Equal(http.StatusForbidden, rec.Code)
		})
	})
}
