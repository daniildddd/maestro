package response

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/errs"
)

var StatusCodeUninitialized = -1

type RWriter struct {
	http.ResponseWriter

	statusCode int
	AppErr     *errs.AppError
	RawErr     error

	UserID   uuid.UUID
	Role     string
	AuthDone bool
}

func NewResponseWriter(w http.ResponseWriter) *RWriter {
	return &RWriter{
		ResponseWriter: w,
		statusCode:     StatusCodeUninitialized,
	}
}

func (rw *RWriter) WriteHeader(statusCode int) {
	if rw.statusCode != StatusCodeUninitialized {
		return
	}

	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *RWriter) Write(b []byte) (int, error) {
	if rw.statusCode == StatusCodeUninitialized {
		rw.statusCode = http.StatusOK
	}

	return rw.ResponseWriter.Write(b)
}

func (rw *RWriter) Written() bool {
	return rw.statusCode != StatusCodeUninitialized
}

func (rw *RWriter) GetStatusCode() int {
	if rw.statusCode == StatusCodeUninitialized {
		return http.StatusOK
	}

	return rw.statusCode
}
