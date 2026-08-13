package response_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/daniildddd/maestro/internal/core/transport/response"
)

func TestNewResponseWriter(t *testing.T) {
	t.Parallel()

	t.Run("wraps underlying ResponseWriter", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		rec := httptest.NewRecorder()
		rw := response.NewResponseWriter(rec)

		_, _ = rw.Write([]byte("hello")) //nolint:errcheck // test-only: httptest ResponseWriter.Write never fails

		is.Equal("hello", rec.Body.String())
	})
}

func TestRWriter_WriteHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
	}{
		{name: "200 OK", statusCode: http.StatusOK},
		{name: "201 Created", statusCode: http.StatusCreated},
		{name: "204 No Content", statusCode: http.StatusNoContent},
		{name: "400 Bad Request", statusCode: http.StatusBadRequest},
		{name: "404 Not Found", statusCode: http.StatusNotFound},
		{name: "500 Internal Server Error", statusCode: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)

			rec := httptest.NewRecorder()
			rw := response.NewResponseWriter(rec)

			rw.WriteHeader(tt.statusCode)

			is.Equal(tt.statusCode, rec.Code)
			is.Equal(tt.statusCode, rw.GetStatusCode())
		})
	}

	t.Run("last WriteHeader wins when called multiple times", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		rec := httptest.NewRecorder()
		rw := response.NewResponseWriter(rec)

		rw.WriteHeader(http.StatusOK)
		rw.WriteHeader(http.StatusNotFound)

		is.Equal(http.StatusNotFound, rw.GetStatusCode())
	})
}

func TestRWriter_GetStatusCode(t *testing.T) {
	t.Parallel()

	t.Run("returns 200 when WriteHeader not called", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		rec := httptest.NewRecorder()
		rw := response.NewResponseWriter(rec)

		is.Equal(http.StatusOK, rw.GetStatusCode())
	})

	t.Run("returns status set by WriteHeader", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		rec := httptest.NewRecorder()
		rw := response.NewResponseWriter(rec)

		rw.WriteHeader(http.StatusCreated)

		is.Equal(http.StatusCreated, rw.GetStatusCode())
	})
}