package server

import (
	"net/http"

	"github.com/daniildddd/maestro/internal/core/transport/middleware"
)

type Route struct {
	Method     string
	Path       string
	Handler    http.HandlerFunc
	Roles      []string
	Middleware []middleware.Middleware
}

func (r *Route) WithMiddleware() http.Handler {
	mw := r.Middleware
	if len(r.Roles) > 0 {
		mw = append(
			[]middleware.Middleware{middleware.RequireRole(r.Roles...)},
			mw...,
		)
	}

	return middleware.ChainMiddleware(
		r.Handler,
		mw...,
	)
}
