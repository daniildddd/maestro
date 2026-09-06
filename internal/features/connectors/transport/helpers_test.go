package transport_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/features/connectors/transport"
)

type errorResponseBody struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Details json.RawMessage `json:"details"`
}

func nopLogger() *logger.Logger {
	return &logger.Logger{Logger: zap.NewNop()}
}

func strPtr(s string) *string { return &s }

func newTestHandler(service transport.ConnectorsService) *transport.ConnectorsHTTPHandler {
	return transport.NewConnectorsHTTPHandler(service)
}

func newConnectorsRequest(t *testing.T, method, path, body string) *http.Request {
	t.Helper()

	var req *http.Request

	if body == "" {
		req = httptest.NewRequest(method, path, http.NoBody)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}

	req = setConnectorsPathValues(req, path)

	return req.WithContext(logger.ToContext(req.Context(), nopLogger()))
}

func setConnectorsPathValues(req *http.Request, path string) *http.Request {
	if path == "/connectors" || path == "/connectors/" {
		return req
	}

	prefixes := []string{"/connectors", "/connector-plugins", "/smt-plugins"}

	for _, prefix := range prefixes {
		if !strings.HasPrefix(path, prefix) {
			continue
		}

		segments := strings.Split(strings.TrimPrefix(path, prefix), "/")

		if len(segments) > 1 {
			req.SetPathValue("id", segments[1])
		}

		if prefix == "/connectors" && len(segments) > 3 {
			req.SetPathValue("task_id", segments[3])
		}

		return req
	}

	return req
}
