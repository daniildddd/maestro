package transport_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/features/audit/transport"
)

func TestAuditHTTPHandler_PrivateRoutes(t *testing.T) {
	t.Parallel()

	is := assert.New(t)
	must := require.New(t)

	handler := transport.NewAuditHTTPHandler(nil)

	routes := handler.PrivateRoutes()

	must.Len(routes, 3)

	type routeKey struct {
		method string
		path   string
	}

	byKey := make(map[routeKey][]string, len(routes))

	for _, r := range routes {
		key := routeKey{method: r.Method, path: r.Path}
		is.NotNil(r.Handler, "handler must be set for %s %s", r.Method, r.Path)
		byKey[key] = r.Roles
	}

	expected := []struct {
		method string
		path   string
		roles  []string
	}{
		{http.MethodGet, "/audit-logs", []string{domain.RoleAdmin}},
		{http.MethodGet, "/audit-logs/{id}", []string{domain.RoleAdmin}},
		{http.MethodDelete, "/audit-logs/{id}", []string{domain.RoleAdmin}},
	}

	must.Len(expected, len(routes))

	for _, e := range expected {
		roles, ok := byKey[routeKey{method: e.method, path: e.path}]
		is.True(ok, "missing route %s %s", e.method, e.path)

		if ok {
			is.Equal(e.roles, roles)
		}
	}
}
