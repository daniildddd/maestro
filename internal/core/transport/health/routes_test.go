package health_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/transport/health"
)

func TestHealthHandler_Routes(t *testing.T) {
	t.Parallel()

	is := assert.New(t)
	must := require.New(t)

	handler := health.NewHealthHandler(nil, nil, health.Config{})

	routes := handler.Routes()

	must.Len(routes, 2)

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
		{method: http.MethodGet, path: "/healthz"},
		{method: http.MethodGet, path: "/readyz"},
	}

	must.Len(expected, len(routes))

	for _, e := range expected {
		is.True(byKey[e], "missing route %s %s", e.method, e.path)
	}
}
