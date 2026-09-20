package transport_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/features/auth/transport"
)

func TestAuthHTTPHandler_PrivateRoutes(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	handler := transport.NewAuthHTTPHandler(nil, transport.Config{})

	routes := handler.PrivateRoutes()

	must.Empty(routes)
}

func TestAuthHTTPHandler_PublicRoutes(t *testing.T) {
	t.Parallel()

	is := assert.New(t)
	must := require.New(t)

	handler := transport.NewAuthHTTPHandler(nil, transport.Config{})

	routes := handler.PublicRoutes()

	must.Len(routes, 3)

	type routeKey struct {
		method string
		path   string
	}

	byKey := make(map[routeKey]bool, len(routes))

	for _, r := range routes {
		key := routeKey{method: r.Method, path: r.Path}
		is.NotNil(r.Handler, "handler must be set for %s %s", r.Method, r.Path)
		is.Empty(r.Roles, "roles must be empty for %s %s", r.Method, r.Path)

		byKey[key] = true
	}

	expected := []routeKey{
		{method: http.MethodPost, path: "/login"},
		{method: http.MethodPost, path: "/logout"},
		{method: http.MethodPost, path: "/refresh"},
	}

	must.Len(expected, len(routes))

	for _, e := range expected {
		is.True(byKey[e], "missing route %s %s", e.method, e.path)
	}
}
