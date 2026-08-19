package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/middleware"
	"github.com/daniildddd/maestro/internal/core/transport/reqctx"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
	"github.com/daniildddd/maestro/internal/core/transport/server"
)

func makeTracingMiddleware(id string) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(id + "->")) //nolint:errcheck // test-only: httptest ResponseWriter.Write never fails
			next.ServeHTTP(w, r)
		})
	}
}

func TestNewAPIVersionRouter(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil router for given routes and version", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		routes := []server.Route{
			{Method: http.MethodGet, Path: "/users", Handler: func(http.ResponseWriter, *http.Request) {}},
		}

		router := server.NewAPIVersionRouter(routes, server.ApiVersion1)

		must.NotNil(router)
	})
}

func TestAPIVersionRouter_Handlers(t *testing.T) {
	t.Parallel()

	t.Run("returns handlers keyed by method and versioned path", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		routes := []server.Route{
			{Method: http.MethodGet, Path: "/users", Handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte("users")) //nolint:errcheck // test-only
			}},
			{Method: http.MethodPost, Path: "/login", Handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte("login")) //nolint:errcheck // test-only
			}},
		}

		router := server.NewAPIVersionRouter(routes, server.ApiVersion1)

		handlers := router.Handlers()

		must.Len(handlers, 2)
		must.Contains(handlers, "GET /api/v1/users")
		must.Contains(handlers, "POST /api/v1/login")

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", http.NoBody)
		rec := httptest.NewRecorder()

		must.Contains(handlers, "GET /api/v1/users")
		handlers["GET /api/v1/users"].ServeHTTP(rec, req)
		must.Equal("users", rec.Body.String())

		req = httptest.NewRequest(http.MethodPost, "/api/v1/login", http.NoBody)
		rec = httptest.NewRecorder()

		must.Contains(handlers, "POST /api/v1/login")
		handlers["POST /api/v1/login"].ServeHTTP(rec, req)
		must.Equal("login", rec.Body.String())
	})

	t.Run("applies router-level middleware around route handler", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		routes := []server.Route{
			{Method: http.MethodGet, Path: "/", Handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte("done")) //nolint:errcheck // test-only
			}},
		}

		router := server.NewAPIVersionRouter(
			routes,
			server.ApiVersion1,
			makeTracingMiddleware("ROUTER1"),
			makeTracingMiddleware("ROUTER2"),
		)

		handlers := router.Handlers()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/", http.NoBody)
		rec := httptest.NewRecorder()

		must.Contains(handlers, "GET /api/v1/")
		handlers["GET /api/v1/"].ServeHTTP(rec, req)

		must.Equal("ROUTER1->ROUTER2->done", rec.Body.String())
	})

	t.Run("applies route-level middleware before router-level", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		routes := []server.Route{
			{
				Method: http.MethodGet,
				Path:   "/",
				Handler: func(w http.ResponseWriter, _ *http.Request) {
					_, _ = w.Write([]byte("done")) //nolint:errcheck // test-only
				},
				Middleware: []middleware.Middleware{makeTracingMiddleware("ROUTE")},
			},
		}

		router := server.NewAPIVersionRouter(
			routes,
			server.ApiVersion1,
			makeTracingMiddleware("ROUTER"),
		)

		handlers := router.Handlers()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/", http.NoBody)
		rec := httptest.NewRecorder()

		must.Contains(handlers, "GET /api/v1/")
		handlers["GET /api/v1/"].ServeHTTP(rec, req)

		must.Equal("ROUTER->ROUTE->done", rec.Body.String())
	})

	t.Run("applies RequireRole from route Roles", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		routes := []server.Route{
			{
				Method: http.MethodGet,
				Path:   "/admin",
				Handler: func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				},
				Roles: []string{"admin"},
			},
		}

		router := server.NewAPIVersionRouter(routes, server.ApiVersion1)

		handlers := router.Handlers()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin", http.NoBody)
		ctx := core_logger.ToContext(req.Context(), nopLogger())
		ctx = reqctx.WithRole(ctx, "user")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		rw := core_http_response.NewResponseWriter(rec)

		must.Contains(handlers, "GET /api/v1/admin")
		handlers["GET /api/v1/admin"].ServeHTTP(rw, req)

		must.Equal(http.StatusForbidden, rec.Code)
	})

	t.Run("returns empty map for no routes", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		router := server.NewAPIVersionRouter(nil, server.ApiVersion1)

		handlers := router.Handlers()

		must.Empty(handlers)
	})

	t.Run("uses different api version in pattern", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		routes := []server.Route{
			{Method: http.MethodGet, Path: "/items", Handler: func(http.ResponseWriter, *http.Request) {}},
		}

		router := server.NewAPIVersionRouter(routes, "v2")

		handlers := router.Handlers()

		must.Len(handlers, 1)
		must.Contains(handlers, "GET /api/v2/items")
	})
}
