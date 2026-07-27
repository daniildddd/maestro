package server

import (
	"net/http"

	"github.com/daniildddd/maestro/internal/core/transport/middleware"
)

type APIVersion string

var ApiVersion1 = APIVersion("v1")

type APIVersionRouter struct {
	routes      []Route
	middlewares []middleware.Middleware
	apiVersion  APIVersion
}

func NewAPIVersionRouter(
	routes []Route,
	apiVersion APIVersion,
	middlewares ...middleware.Middleware,
) *APIVersionRouter {
	return &APIVersionRouter{
		routes:      routes,
		middlewares: middlewares,
		apiVersion:  apiVersion,
	}
}

func (a *APIVersionRouter) Handlers() map[string]http.Handler {
	handlers := make(map[string]http.Handler, len(a.routes))

	for _, route := range a.routes {
		pattern := route.Method + " /api/" + string(a.apiVersion) + route.Path
		handler := middleware.ChainMiddleware(route.WithMiddleware(), a.middlewares...)
		handlers[pattern] = handler
	}

	return handlers
}
