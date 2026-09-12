package health

import (
	"context"
	"net/http"

	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
	"github.com/daniildddd/maestro/internal/core/transport/server"
)

type connectChecker interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	pool    core_postgres_pool.Pool
	connect connectChecker
	cfg     Config
}

func NewHealthHandler(
	pool core_postgres_pool.Pool,
	connect connectChecker,
	cfg Config,
) *HealthHandler {
	return &HealthHandler{pool: pool, connect: connect, cfg: cfg}
}

func (h *HealthHandler) Routes() []server.Route {
	return []server.Route{
		{
			Method:  http.MethodGet,
			Path:    "/healthz",
			Handler: h.Live,
		},
		{
			Method:  http.MethodGet,
			Path:    "/readyz",
			Handler: h.Ready,
		},
	}
}

func (h *HealthHandler) Live(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.ReadyTimeout)
	defer cancel()

	var one int

	if err := h.pool.QueryRow(ctx, "SELECT 1").Scan(&one); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)

		return
	}

	if err := h.connect.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
