package response_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/transport/response"
)

func TestNewResponseWriter(t *testing.T) {
	t.Parallel()

	t.Run("returns initialized RWriter wrapping the given ResponseWriter", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		rec := httptest.NewRecorder()
		rw := response.NewResponseWriter(rec)

		must.NotNil(rw)
		must.Same(rec, rw.ResponseWriter)
		must.Nil(rw.AppErr)
		must.NoError(rw.RawErr)
		must.False(rw.Written())
	})
}

func TestRWriter_WriteHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
	}{
		{
			name:       "200 status passes through",
			statusCode: http.StatusOK,
		},
		{
			name:       "201 status passes through",
			statusCode: http.StatusCreated,
		},
		{
			name:       "204 status passes through",
			statusCode: http.StatusNoContent,
		},
		{
			name:       "400 status passes through",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "404 status passes through",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "500 status passes through",
			statusCode: http.StatusInternalServerError,
		},
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
			is.True(rw.Written())
		})
	}

	t.Run("first WriteHeader wins when called multiple times", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		rec := httptest.NewRecorder()
		rw := response.NewResponseWriter(rec)

		rw.WriteHeader(http.StatusOK)
		rw.WriteHeader(http.StatusNotFound)

		is.Equal(http.StatusOK, rec.Code)
		is.Equal(http.StatusOK, rw.GetStatusCode())
		is.True(rw.Written())
	})
}

func TestRWriter_Write(t *testing.T) {
	t.Parallel()

	t.Run("write without WriteHeader records implicit 200", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		rec := httptest.NewRecorder()
		rw := response.NewResponseWriter(rec)

		//nolint:errcheck // test-only: httptest ResponseWriter.Write never fails
		_, err := rw.Write([]byte("hello"))
		must.NoError(err)
		is.Equal(http.StatusOK, rec.Code)
		is.Equal(http.StatusOK, rw.GetStatusCode())
		is.True(rw.Written())
	})

	t.Run("write after WriteHeader does not change status", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		rec := httptest.NewRecorder()
		rw := response.NewResponseWriter(rec)

		rw.WriteHeader(http.StatusInternalServerError)

		//nolint:errcheck // test-only: httptest ResponseWriter.Write never fails
		_, err := rw.Write([]byte("boom"))
		must.NoError(err)
		is.Equal(http.StatusInternalServerError, rec.Code)
		is.Equal(http.StatusInternalServerError, rw.GetStatusCode())
	})

	t.Run("WriteHeader after write does not change status", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		rec := httptest.NewRecorder()
		rw := response.NewResponseWriter(rec)

		//nolint:errcheck // test-only: httptest ResponseWriter.Write never fails
		_, err := rw.Write([]byte("hello"))
		must.NoError(err)

		rw.WriteHeader(http.StatusInternalServerError)

		is.Equal(http.StatusOK, rec.Code)
		is.Equal(http.StatusOK, rw.GetStatusCode())
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
