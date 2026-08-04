package response

import (
	"net/http"

	"github.com/daniildddd/maestro/internal/core/errs"
)

var StatusCodeUninitialized = -1

type RWriter struct {
	http.ResponseWriter

	statusCode int
	AppErr     *errs.AppError
	RawErr     error
}

func NewResponseWriter(w http.ResponseWriter) *RWriter {
	return &RWriter{
		ResponseWriter: w,
		statusCode:     StatusCodeUninitialized,
	}
}

func (rw *RWriter) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
	rw.statusCode = statusCode
}

func (rw *RWriter) GetStatusCode() int {
	if rw.statusCode == StatusCodeUninitialized {
		return http.StatusOK
	}

	return rw.statusCode
}
