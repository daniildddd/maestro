package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/middleware"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

func TestTrace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		method      string
		path        string
		observerLvl zapcore.Level
		handler     func(w http.ResponseWriter, log *logger.Logger)
		wantStatus  int
		wantLogs    int
		wantLevel   zapcore.Level
		wantMsg     string
		wantCode    string
		wantError   bool
	}{
		{
			name:        "success: 200 OK logged at info",
			method:      http.MethodGet,
			path:        "/api/v1/users",
			observerLvl: zapcore.InfoLevel,
			handler: func(w http.ResponseWriter, _ *logger.Logger) {
				w.WriteHeader(http.StatusOK)
			},
			wantStatus: http.StatusOK,
			wantLogs:   1,
			wantLevel:  zapcore.InfoLevel,
			wantMsg:    "request completed successfully",
		},
		{
			name:        "success: 204 No Content status reflected in log",
			method:      http.MethodGet,
			path:        "/api/v1/users/42",
			observerLvl: zapcore.InfoLevel,
			handler: func(w http.ResponseWriter, _ *logger.Logger) {
				w.WriteHeader(http.StatusNoContent)
			},
			wantStatus: http.StatusNoContent,
			wantLogs:   1,
			wantLevel:  zapcore.InfoLevel,
			wantMsg:    "request completed successfully",
		},
		{
			name:        "success: 404 without AppErr uses success path",
			method:      http.MethodGet,
			path:        "/api/v1/missing",
			observerLvl: zapcore.InfoLevel,
			handler: func(w http.ResponseWriter, _ *logger.Logger) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantLogs:   1,
			wantLevel:  zapcore.InfoLevel,
			wantMsg:    "request completed successfully",
		},
		{
			name:        "success: POST method and path recorded in fields",
			method:      http.MethodPost,
			path:        "/api/v1/login",
			observerLvl: zapcore.InfoLevel,
			handler: func(w http.ResponseWriter, _ *logger.Logger) {
				w.WriteHeader(http.StatusCreated)
			},
			wantStatus: http.StatusCreated,
			wantLogs:   1,
			wantLevel:  zapcore.InfoLevel,
			wantMsg:    "request completed successfully",
		},
		{
			name:        "error: ErrInvalidCredentials logged at info level",
			method:      http.MethodPost,
			path:        "/api/v1/login",
			observerLvl: zapcore.InfoLevel,
			handler: func(w http.ResponseWriter, log *logger.Logger) {
				rh := core_http_response.NewHTTPResponseHandler(w, log)
				rh.ErrorResponse(fmt.Errorf("service: %w", errs.ErrInvalidCredentials))
			},
			wantStatus: http.StatusBadRequest,
			wantLogs:   1,
			wantLevel:  zapcore.InfoLevel,
			wantMsg:    "request completed with error",
			wantCode:   "INVALID_CREDENTIALS",
			wantError:  true,
		},
		{
			name:        "error: ErrInvalidRequestBody logged at warn level",
			method:      http.MethodPost,
			path:        "/api/v1/users",
			observerLvl: zapcore.WarnLevel,
			handler: func(w http.ResponseWriter, log *logger.Logger) {
				rh := core_http_response.NewHTTPResponseHandler(w, log)
				rh.ErrorResponse(errs.ErrInvalidRequestBody)
			},
			wantStatus: http.StatusBadRequest,
			wantLogs:   1,
			wantLevel:  zapcore.WarnLevel,
			wantMsg:    "request completed with error",
			wantCode:   "INVALID_REQUEST_BODY",
			wantError:  true,
		},
		{
			name:        "error: ErrInternal logged at error level",
			method:      http.MethodGet,
			path:        "/api/v1/crash",
			observerLvl: zapcore.ErrorLevel,
			handler: func(w http.ResponseWriter, log *logger.Logger) {
				rh := core_http_response.NewHTTPResponseHandler(w, log)
				rh.ErrorResponse(errs.ErrInternal)
			},
			wantStatus: http.StatusInternalServerError,
			wantLogs:   1,
			wantLevel:  zapcore.ErrorLevel,
			wantMsg:    "request completed with error",
			wantCode:   "INTERNAL_ERROR",
			wantError:  true,
		},
		{
			name:        "level filter: error below observer level not logged",
			method:      http.MethodPost,
			path:        "/api/v1/login",
			observerLvl: zapcore.WarnLevel,
			handler: func(w http.ResponseWriter, log *logger.Logger) {
				rh := core_http_response.NewHTTPResponseHandler(w, log)
				rh.ErrorResponse(fmt.Errorf("service: %w", errs.ErrInvalidCredentials))
			},
			wantStatus: http.StatusBadRequest,
			wantLogs:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)

			log, recordedLogs := newObservableLogger(t, tt.observerLvl)

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				tt.handler(w, log)
			})

			req := newTestRequest(t, tt.method, tt.path, nil)
			req = req.WithContext(logger.ToContext(req.Context(), log))
			rec := httptest.NewRecorder()

			chained := middleware.Trace()(nextHandler)
			chained.ServeHTTP(rec, req)

			must.Equal(tt.wantStatus, rec.Code)

			logs := recordedLogs.All()
			must.Len(logs, tt.wantLogs)

			if tt.wantLogs == 0 {
				return
			}

			must.Equal(tt.wantLevel, logs[0].Level)
			must.Contains(logs[0].Message, tt.wantMsg)

			ctxMap := logs[0].ContextMap()
			must.Equal(tt.method, ctxMap["method"])
			must.Equal(tt.path, ctxMap["path"])
			must.Equal(int64(tt.wantStatus), ctxMap["status"])

			latency, ok := ctxMap["latency"].(time.Duration)
			must.True(ok, "latency field must be time.Duration")
			must.GreaterOrEqual(latency, time.Duration(0), "latency must be non-negative")

			if tt.wantError {
				must.Equal(tt.wantCode, ctxMap["code"])
				must.NotNil(ctxMap["error"])

				return
			}

			_, hasCode := ctxMap["code"]
			must.False(hasCode, "success path must not log code field")

			_, hasErr := ctxMap["error"]
			must.False(hasErr, "success path must not log error field")
		})
	}
}
