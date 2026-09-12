package health_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
	"github.com/daniildddd/maestro/internal/core/transport/health"
)

type stubRow struct {
	err error
}

func (r stubRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}

	if len(dest) > 0 {
		if p, ok := dest[0].(*int); ok {
			*p = 1
		}
	}

	return nil
}

type stubPool struct {
	core_postgres_pool.Pool

	row   stubRow
	calls int
}

func (s *stubPool) QueryRow(_ context.Context, _ string, _ ...any) core_postgres_pool.Row {
	s.calls++

	return s.row
}

type stubConnect struct {
	err         error
	calls       int
	deadline    time.Time
	hasDeadline bool
}

func (s *stubConnect) Ping(ctx context.Context) error {
	s.calls++
	s.deadline, s.hasDeadline = ctx.Deadline()

	return s.err
}

func TestNewHealthHandler(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	handler := health.NewHealthHandler(&stubPool{}, &stubConnect{}, health.Config{ReadyTimeout: 2 * time.Second})

	must.NotNil(handler)
}

func TestHealthHandler_Live(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	handler := health.NewHealthHandler(&stubPool{}, &stubConnect{}, health.Config{ReadyTimeout: 2 * time.Second})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", http.NoBody)

	handler.Live(rec, req)

	must.Equal(http.StatusNoContent, rec.Code)
	must.Empty(rec.Body.String())
}

func TestHealthHandler_Ready(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		scanErr    error
		pingErr    error
		wantStatus int
		wantPing   int
	}{
		{
			name:       "healthy dependencies return no content",
			scanErr:    nil,
			pingErr:    nil,
			wantStatus: http.StatusNoContent,
			wantPing:   1,
		},
		{
			name:       "database failure returns service unavailable",
			scanErr:    errors.New("connection refused"),
			pingErr:    nil,
			wantStatus: http.StatusServiceUnavailable,
			wantPing:   0,
		},
		{
			name:       "connect failure returns service unavailable",
			scanErr:    nil,
			pingErr:    errors.New("connection refused"),
			wantStatus: http.StatusServiceUnavailable,
			wantPing:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			pool := &stubPool{row: stubRow{err: tt.scanErr}}
			connect := &stubConnect{err: tt.pingErr}
			handler := health.NewHealthHandler(pool, connect, health.Config{ReadyTimeout: 5 * time.Second})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/readyz", http.NoBody)

			handler.Ready(rec, req)

			must.Equal(tt.wantStatus, rec.Code)
			must.Empty(rec.Body.String())

			must.Equal(1, pool.calls)
			must.Equal(tt.wantPing, connect.calls)

			if tt.wantPing > 0 {
				must.True(connect.hasDeadline)
				must.WithinDuration(time.Now().Add(5*time.Second), connect.deadline, 2*time.Second)
			}
		})
	}
}
