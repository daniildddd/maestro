package response_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/stretchr/testify/require"

	core_errs "github.com/daniildddd/maestro/internal/core/errs"
	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/response"
)

func nopLogger() *core_logger.Logger {
	return &core_logger.Logger{Logger: zap.NewNop()}
}

func newObservableLogger(t *testing.T, lvl zapcore.Level) (*core_logger.Logger, *observer.ObservedLogs) {
	t.Helper()

	core, recorder := observer.New(lvl)

	return &core_logger.Logger{Logger: zap.New(core)}, recorder
}

func TestNewHTTPResponseHandler(t *testing.T) {
	t.Parallel()
	must := require.New(t)

	rec := httptest.NewRecorder()
	h := response.NewHTTPResponseHandler(rec, nopLogger())

	must.NotNil(h)
}

func TestHTTPResponseHandler_ErrorResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		err            error
		wantStatus     int
		wantCode       string
		wantMessage    string
		wantWWWAuth    bool
		wantAppErrCode string
	}{
		{
			name:           "AppError 401 sets WWW-Authenticate and writes error body",
			err:            core_errs.ErrAccessTokenMissing,
			wantStatus:     http.StatusUnauthorized,
			wantCode:       "ACCESS_TOKEN_MISSING",
			wantMessage:    "Access token missing",
			wantWWWAuth:    true,
			wantAppErrCode: "ACCESS_TOKEN_MISSING",
		},
		{
			name:           "AppError 400 does not set WWW-Authenticate",
			err:            core_errs.ErrInvalidCredentials,
			wantStatus:     http.StatusBadRequest,
			wantCode:       "INVALID_CREDENTIALS",
			wantMessage:    "Invalid credentials",
			wantWWWAuth:    false,
			wantAppErrCode: "INVALID_CREDENTIALS",
		},
		{
			name:           "AppError 500 writes internal error body",
			err:            core_errs.ErrInternal,
			wantStatus:     http.StatusInternalServerError,
			wantCode:       "INTERNAL_ERROR",
			wantMessage:    "Internal server error",
			wantWWWAuth:    false,
			wantAppErrCode: "INTERNAL_ERROR",
		},
		{
			name:           "wrapped AppError unwraps via errors.As",
			err:            fmt.Errorf("service: %w", core_errs.ErrInvalidRequestBody),
			wantStatus:     http.StatusBadRequest,
			wantCode:       "INVALID_REQUEST_BODY",
			wantMessage:    "Invalid request body",
			wantWWWAuth:    false,
			wantAppErrCode: "INVALID_REQUEST_BODY",
		},
		{
			name:           "non-AppError falls back to ErrInternal",
			err:            errors.New("something went wrong"),
			wantStatus:     http.StatusInternalServerError,
			wantCode:       "INTERNAL_ERROR",
			wantMessage:    "Internal server error",
			wantWWWAuth:    false,
			wantAppErrCode: "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)

			rec := httptest.NewRecorder()
			rw := response.NewResponseWriter(rec)
			h := response.NewHTTPResponseHandler(rw, nopLogger())

			h.ErrorResponse(tt.err)

			must.Equal(tt.wantStatus, rec.Code)
			must.Equal("application/json", rec.Header().Get("Content-Type"))

			if tt.wantWWWAuth {
				must.Equal("Bearer", rec.Header().Get("WWW-Authenticate"))
			} else {
				must.Empty(rec.Header().Get("WWW-Authenticate"))
			}

			var body response.ErrorResponse

			must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
			must.Equal(tt.wantCode, body.Code)
			must.Equal(tt.wantMessage, body.Message)

			must.Equal(tt.wantAppErrCode, rw.AppErr.Code)
			must.Error(rw.RawErr)
		})
	}

	t.Run("panics when underlying writer is not *RWriter", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		rec := httptest.NewRecorder()
		h := response.NewHTTPResponseHandler(rec, nopLogger())

		must.Panics(func() {
			h.ErrorResponse(core_errs.ErrInternal)
		})
	})
}

func TestHTTPResponseHandler_NoContent(t *testing.T) {
	t.Parallel()
	must := require.New(t)

	rec := httptest.NewRecorder()
	h := response.NewHTTPResponseHandler(rec, nopLogger())

	h.NoContent()

	must.Equal(http.StatusNoContent, rec.Code)
}

func TestHTTPResponseHandler_JSONResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       any
		statusCode int
	}{
		{
			name:       "writes ErrorResponse struct as JSON with 200",
			body:       response.ErrorResponse{Code: "TEST_CODE", Message: "test message"},
			statusCode: http.StatusOK,
		},
		{
			name:       "writes map as JSON with 201",
			body:       map[string]string{"id": "123"},
			statusCode: http.StatusCreated,
		},
		{
			name:       "writes ErrorResponse with 500",
			body:       response.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Internal server error"},
			statusCode: http.StatusInternalServerError,
		},
		{
			name:       "nil body encodes as null",
			body:       nil,
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)

			rec := httptest.NewRecorder()
			h := response.NewHTTPResponseHandler(rec, nopLogger())

			h.JSONResponse(tt.body, tt.statusCode)

			must.Equal(tt.statusCode, rec.Code)
			must.Equal("application/json", rec.Header().Get("Content-Type"))

			expected, err := json.Marshal(tt.body)
			must.NoError(err)
			must.Equal(string(expected), strings.TrimSpace(rec.Body.String()))
		})
	}

	t.Run("encode error logs at error level with error field", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		log, recordedLogs := newObservableLogger(t, zapcore.ErrorLevel)
		rec := httptest.NewRecorder()
		h := response.NewHTTPResponseHandler(rec, log)

		h.JSONResponse(make(chan int), http.StatusOK)

		must.Equal(http.StatusOK, rec.Code)
		must.Equal("application/json", rec.Header().Get("Content-Type"))

		logs := recordedLogs.All()
		must.Len(logs, 1)
		must.Equal(zapcore.ErrorLevel, logs[0].Level)
		must.Equal("write HTTP response", logs[0].Message)

		ctxMap := logs[0].ContextMap()
		must.NotEmpty(ctxMap["error"], "error field must not be empty")
	})
}

func TestHTTPResponseHandler_PanicResponse(t *testing.T) {
	t.Parallel()

	t.Run("writes 500 with internal error for string panic", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		rec := httptest.NewRecorder()
		rw := response.NewResponseWriter(rec)
		h := response.NewHTTPResponseHandler(rw, nopLogger())

		h.PanicResponse("something went wrong")

		must.Equal(http.StatusInternalServerError, rec.Code)
		must.Equal("application/json", rec.Header().Get("Content-Type"))

		var body response.ErrorResponse

		must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
		must.Equal("INTERNAL_ERROR", body.Code)
		must.Equal("Internal server error", body.Message)

		must.Equal(core_errs.ErrInternal.Code, rw.AppErr.Code)
		must.Contains(rw.RawErr.Error(), "unexpected panic")
		must.Contains(rw.RawErr.Error(), "something went wrong")
	})

	t.Run("panics when underlying writer is not *RWriter", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		rec := httptest.NewRecorder()
		h := response.NewHTTPResponseHandler(rec, nopLogger())

		must.Panics(func() {
			h.PanicResponse("boom")
		})
	})
}
