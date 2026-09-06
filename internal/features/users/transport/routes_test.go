package transport_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
)

const usersByIDPath = "/users/{id}"

type routeSpec struct {
	roles []string
}

func TestUsersHTTPHandlerPrivateRoutes(t *testing.T) {
	t.Parallel()

	is := assert.New(t)
	must := require.New(t)

	handler := newUsersTestHandler(NewMockUsersService(t))

	routes := handler.PrivateRoutes()

	must.Len(routes, 9)

	type routeKey struct {
		method string
		path   string
	}

	byKey := make(map[routeKey]routeSpec, len(routes))

	for _, r := range routes {
		key := routeKey{method: r.Method, path: r.Path}

		is.NotEmpty(r.Handler, "handler must be set for %s %s", r.Method, r.Path)
		byKey[key] = routeSpec{roles: r.Roles}
	}

	expected := []struct {
		method string
		path   string
		roles  []string
	}{
		{http.MethodPost, "/users", []string{domain.RoleAdmin}},
		{http.MethodGet, "/users", []string{domain.RoleAdmin}},
		{http.MethodPatch, "/users/{id}/password", []string{domain.RoleAdmin}},
		{http.MethodPatch, usersByIDPath, []string{domain.RoleAdmin}},
		{http.MethodGet, usersByIDPath, []string{domain.RoleUser, domain.RoleAdmin}},
		{http.MethodDelete, usersByIDPath, []string{domain.RoleAdmin}},
		{http.MethodGet, "/users/me", []string{domain.RoleUser, domain.RoleAdmin}},
		{http.MethodDelete, "/users/me", []string{domain.RoleUser, domain.RoleAdmin}},
		{http.MethodPatch, "/users/me/password", []string{domain.RoleUser, domain.RoleAdmin}},
	}

	must.Len(expected, len(routes))

	for _, e := range expected {
		spec, ok := byKey[routeKey{method: e.method, path: e.path}]
		is.True(ok, "missing route %s %s", e.method, e.path)

		if ok {
			is.Equal(e.roles, spec.roles)
		}
	}
}
