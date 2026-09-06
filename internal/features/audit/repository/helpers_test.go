package repository_test

import (
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/daniildddd/maestro/internal/core/domain"
	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/features/audit/repository"
)

type stubCommandTag struct {
	affected int64
}

func (s stubCommandTag) RowsAffected() int64 { return s.affected }

func nopLogger() *core_logger.Logger {
	return &core_logger.Logger{Logger: zap.NewNop()}
}

func newAuditRepo(pool *MockPool) *repository.AuditRepository {
	return repository.NewAuditRepository(pool)
}

func strPtr(s string) *string { return &s }

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
		ID:            id,
		Action:        domain.ActionAuthLogin,
		Outcome:       domain.OutcomeSuccess,
		FailureReason: "invalid_credentials",
		ActorID:       actorID,
		ActorLogin:    actorLogin,
		SubjectType:   domain.AuditSubjectUser,
		SubjectID:     subjectID,
		SubjectName:   subjectName,
		StateBefore:   map[string]any{"role": "user"},
		StateAfter:    map[string]any{"role": "admin"},
		RequestID:     requestID,
		IP:            ip,
		UserAgent:     userAgent,
		CreatedAt:     createdAt,
	}
}

type auditRowFixture struct {
	id            uuid.UUID
	action        string
	outcome       string
	failureReason *string
	actorID       *uuid.UUID
	actorLogin    *string
	subjectType   string
	subjectID     *string
	subjectName   *string
	stateBefore   []byte
	stateAfter    []byte
	requestID     *string
	ip            *string
	userAgent     *string
	createdAt     time.Time
}

func mustAuditRow(id uuid.UUID) auditRowFixture {
	actorID := id

	return auditRowFixture{
		id:            id,
		action:        "auth.login",
		outcome:       "success",
		failureReason: strPtr("invalid_credentials"),
		actorID:       &actorID,
		actorLogin:    strPtr("alice"),
		subjectType:   "user",
		subjectID:     strPtr(id.String()),
		subjectName:   strPtr("alice"),
		stateBefore:   []byte(`{"role":"user"}`),
		stateAfter:    []byte(`{"role":"admin"}`),
		requestID:     strPtr("req-" + id.String()),
		ip:            strPtr("203.0.113.7"),
		userAgent:     strPtr("test-agent"),
		createdAt:     time.Now().UTC().Truncate(time.Second),
	}
}

//nolint:unparam // test helper mirrors the audit_logs table row; any argument may vary per test
func scanEventIntoRow(row *MockRow, fixture auditRowFixture) {
	row.EXPECT().
		Scan(mock.Anything).
		Run(func(dest ...any) {
			fillEventDest(dest, fixture)
		}).
		Return(nil).
		Once()
}

//nolint:unparam // test helper mirrors the audit_logs table rows; any argument may vary per test
func scanEventIntoRows(rows *MockRows, fixture auditRowFixture) {
	rows.EXPECT().
		Scan(mock.Anything).
		Run(func(dest ...any) {
			fillEventDest(dest, fixture)
		}).
		Return(nil).
		Once()
}

func fillEventDest(ptrs []any, row auditRowFixture) {
	fillEventCoreDest(ptrs, row)
	fillEventStateDest(ptrs, row)
}

func fillEventCoreDest(ptrs []any, row auditRowFixture) {
	if idPtr, ok := ptrs[0].(*uuid.UUID); ok {
		*idPtr = row.id
	}

	if actionPtr, ok := ptrs[1].(*string); ok {
		*actionPtr = row.action
	}

	if outcomePtr, ok := ptrs[2].(*string); ok {
		*outcomePtr = row.outcome
	}

	if failureReasonPP, ok := ptrs[3].(**string); ok {
		*failureReasonPP = row.failureReason
	}

	if actorIDPP, ok := ptrs[4].(**uuid.UUID); ok {
		*actorIDPP = row.actorID
	}

	if actorLoginPP, ok := ptrs[5].(**string); ok {
		*actorLoginPP = row.actorLogin
	}

	if subjectTypePtr, ok := ptrs[6].(*string); ok {
		*subjectTypePtr = row.subjectType
	}

	if subjectIDPP, ok := ptrs[7].(**string); ok {
		*subjectIDPP = row.subjectID
	}

	if subjectNamePP, ok := ptrs[8].(**string); ok {
		*subjectNamePP = row.subjectName
	}
}

func fillEventStateDest(ptrs []any, row auditRowFixture) {
	if stateBeforePtr, ok := ptrs[9].(*[]byte); ok {
		*stateBeforePtr = row.stateBefore
	}

	if stateAfterPtr, ok := ptrs[10].(*[]byte); ok {
		*stateAfterPtr = row.stateAfter
	}

	if requestIDPP, ok := ptrs[11].(**string); ok {
		*requestIDPP = row.requestID
	}

	if ipPP, ok := ptrs[12].(**string); ok {
		*ipPP = row.ip
	}

	if userAgentPP, ok := ptrs[13].(**string); ok {
		*userAgentPP = row.userAgent
	}

	if createdAtPtr, ok := ptrs[14].(*time.Time); ok {
		*createdAtPtr = row.createdAt
	}
}
