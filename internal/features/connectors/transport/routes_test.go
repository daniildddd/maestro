package transport_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/features/connectors/transport"
)

func TestConnectorsHTTPHandler_PrivateRoutes(t *testing.T) {
	t.Parallel()

	is := assert.New(t)
	must := require.New(t)

	handler := transport.NewConnectorsHTTPHandler(nil)

	routes := handler.PrivateRoutes()

	must.Len(routes, 15)

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

	roles := []string{domain.RoleUser, domain.RoleAdmin}

	expected := []struct {
		method string
		path   string
		roles  []string
	}{
		{http.MethodGet, "/connectors", roles},
		{http.MethodPost, "/connectors", roles},
		{http.MethodPost, "/connectors/validate", roles},
		{http.MethodGet, "/connectors/{id}", roles},
		{http.MethodDelete, "/connectors/{id}", roles},
		{http.MethodPatch, "/connectors/{id}", roles},
		{http.MethodGet, "/connectors/{id}/tasks/{task_id}", roles},
		{http.MethodPost, "/connectors/{id}/tasks/{task_id}/restart", roles},
		{http.MethodPost, "/connectors/{id}/pause", roles},
		{http.MethodPost, "/connectors/{id}/resume", roles},
		{http.MethodPost, "/connectors/{id}/restart", roles},
		{http.MethodGet, "/connector-plugins", roles},
		{http.MethodGet, "/connector-plugins/{id}/schema", roles},
		{http.MethodGet, "/smt-plugins", roles},
		{http.MethodGet, "/smt-plugins/{id}/schema", roles},
	}

	must.Len(expected, len(routes))

	for _, e := range expected {
		got, ok := byKey[routeKey{method: e.method, path: e.path}]
		is.True(ok, "missing route %s %s", e.method, e.path)

		if ok {
			is.Equal(e.roles, got)
		}
	}
}
