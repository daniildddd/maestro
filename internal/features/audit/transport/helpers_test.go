package transport_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/features/audit/transport"
)

func nopLogger() *core_logger.Logger {
	return &core_logger.Logger{Logger: zap.NewNop()}
}

func newAuditTestHandler(auditService transport.AuditService) *transport.AuditHTTPHandler {
	return transport.NewAuditHTTPHandler(auditService)
}

type errorResponseBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func mustAuditEvent(t *testing.T, id uuid.UUID) domain.AuditEvent {
	t.Helper()

	createdAt := time.Now().UTC().Truncate(time.Second)
	actorID := id
	actorLogin := "alice"
	subjectID := id.String()
	subjectName := "alice"
	requestID := "req-" + id.String()
	ip := "203.0.113.7"
	userAgent := "test-agent"

	return domain.AuditEvent{
		ID:          id,
		Action:      domain.ActionAuthLogin,
		Outcome:     domain.OutcomeSuccess,
		ActorID:     actorID,
		ActorLogin:  actorLogin,
		SubjectType: domain.AuditSubjectUser,
		SubjectID:   subjectID,
		SubjectName: subjectName,
		RequestID:   requestID,
		IP:          ip,
		UserAgent:   userAgent,
		CreatedAt:   createdAt,
	}
}

func newAuditRequest(t *testing.T, method, path string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, path, http.NoBody)

	if path != "/audit-logs" && strings.HasPrefix(path, "/audit-logs/") {
		idSegment := strings.Split(strings.TrimPrefix(path, "/audit-logs/"), "/")[0]
		req.SetPathValue("id", idSegment)
	}

	return req.WithContext(core_logger.ToContext(req.Context(), nopLogger()))
}
