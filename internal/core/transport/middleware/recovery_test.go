package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/require"

	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/middleware"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func newRequestWithLogger(t *testing.T) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	testLog := &core_logger.Logger{Logger: zap.NewNop()}
	ctx := core_logger.ToContext(req.Context(), testLog)

	return req.WithContext(ctx)
}

func TestRecovery(t *testing.T) {
	t.Parallel()
	t.Run("no panic passes through", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("done")) //nolint:errcheck // test-only: httptest ResponseWriter.Write never fails
		})

		chained := middleware.Recovery()(nextHandler)

		req := newRequestWithLogger(t)
		rec := httptest.NewRecorder()
		rw := core_http_response.NewResponseWriter(rec)

		chained.ServeHTTP(rw, req)
		must.Equal(http.StatusOK, rec.Code)
		must.Equal("done", rec.Body.String())
	})

	t.Run("panic recovers as 500", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)
		nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			panic("something went horribly wrong")
		})

		chained := middleware.Recovery()(nextHandler)

		req := newRequestWithLogger(t)
		rec := httptest.NewRecorder()
		rw := core_http_response.NewResponseWriter(rec)

		chained.ServeHTTP(rw, req)
		must.Equal(http.StatusInternalServerError, rec.Code)
		must.Contains(rec.Body.String(), "INTERNAL_ERROR")
	})
}
