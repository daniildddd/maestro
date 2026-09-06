package repository_test

import (
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/daniildddd/maestro/internal/core/domain"
	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/features/audit/repository"
)

func nopLogger() *core_logger.Logger {
	return &core_logger.Logger{Logger: zap.NewNop()}
}

func newAuditRepo(pool *MockPool) *repository.AuditRepository {
	return repository.NewAuditRepository(pool, nopLogger())
}

type stubCommandTag struct {
	affected int64
}

func (s stubCommandTag) RowsAffected() int64 { return s.affected }

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
		RequestID:     requestID,
		IP:            ip,
		UserAgent:     userAgent,
		CreatedAt:     createdAt,
	}
}

//nolint:unparam // test helper mirrors the audit_logs table row; any argument may vary per test
func scanEventIntoRow(row *MockRow, event domain.AuditEvent) {
	row.EXPECT().
		Scan(mock.Anything).
		Run(func(dest ...any) {
			ptrs := dest

			if len(dest) == 1 {
				if inner, ok := dest[0].([]any); ok {
					ptrs = inner
				}
			}

			fillEventDest(ptrs, event)
		}).
		Return(nil).
		Once()
}

//nolint:unparam // test helper mirrors the audit_logs table rows; any argument may vary per test
func scanEventIntoRows(rows *MockRows, event domain.AuditEvent) {
	rows.EXPECT().
		Scan(mock.Anything).
		Run(func(dest ...any) {
			ptrs := dest

			if len(dest) == 1 {
				if inner, ok := dest[0].([]any); ok {
					ptrs = inner
				}
			}

			fillEventDest(ptrs, event)
		}).
		Return(nil).
		Once()
}

func fillEventDest(ptrs []any, event domain.AuditEvent) {
	fillEventCoreDest(ptrs, event)
	fillEventStateDest(ptrs, event)
}

func fillEventCoreDest(ptrs []any, event domain.AuditEvent) {
	if idPtr, ok := ptrs[0].(*uuid.UUID); ok {
		*idPtr = event.ID
	}

	if actionPtr, ok := ptrs[1].(*string); ok {
		*actionPtr = string(event.Action)
	}

	if outcomePtr, ok := ptrs[2].(*string); ok {
		*outcomePtr = string(event.Outcome)
	}

	if failureReasonPP, ok := ptrs[3].(**string); ok {
		if event.FailureReason != "" {
			val := event.FailureReason
			*failureReasonPP = &val
		}
	}

	if actorIDPP, ok := ptrs[4].(**uuid.UUID); ok {
		if event.ActorID != uuid.Nil {
			val := event.ActorID
			*actorIDPP = &val
		}
	}

	if actorLoginPP, ok := ptrs[5].(**string); ok {
		if event.ActorLogin != "" {
			val := event.ActorLogin
			*actorLoginPP = &val
		}
	}

	if subjectTypePtr, ok := ptrs[6].(*string); ok {
		*subjectTypePtr = event.SubjectType
	}

	if subjectIDPP, ok := ptrs[7].(**string); ok {
		if event.SubjectID != "" {
			val := event.SubjectID
			*subjectIDPP = &val
		}
	}

	if subjectNamePP, ok := ptrs[8].(**string); ok {
		if event.SubjectName != "" {
			val := event.SubjectName
			*subjectNamePP = &val
		}
	}
}

func fillEventStateDest(ptrs []any, event domain.AuditEvent) {
	if stateBeforePtr, ok := ptrs[9].(*[]byte); ok {
		if event.StateBefore != nil {
			encoded, err := json.Marshal(event.StateBefore)
			if err != nil {
				panic("unserializable test state_before: " + err.Error())
			}

			*stateBeforePtr = encoded
		}
	}

	if stateAfterPtr, ok := ptrs[10].(*[]byte); ok {
		if event.StateAfter != nil {
			encoded, err := json.Marshal(event.StateAfter)
			if err != nil {
				panic("unserializable test state_after: " + err.Error())
			}

			*stateAfterPtr = encoded
		}
	}

	if requestIDPP, ok := ptrs[11].(**string); ok {
		if event.RequestID != "" {
			val := event.RequestID
			*requestIDPP = &val
		}
	}

	if ipPP, ok := ptrs[12].(**string); ok {
		if event.IP != "" {
			val := event.IP
			*ipPP = &val
		}
	}

	if userAgentPP, ok := ptrs[13].(**string); ok {
		if event.UserAgent != "" {
			val := event.UserAgent
			*userAgentPP = &val
		}
	}

	if createdAtPtr, ok := ptrs[14].(*time.Time); ok {
		*createdAtPtr = event.CreatedAt
	}
}

func corruptStateEvent(t *testing.T, id uuid.UUID) domain.AuditEvent {
	t.Helper()

	event := mustAuditEvent(t, id)
	event.StateBefore = map[string]any{"state": "broken"}

	return event
}

func scanCorruptStateIntoRow(row *MockRow, event domain.AuditEvent) {
	row.EXPECT().
		Scan(mock.Anything).
		Run(func(dest ...any) {
			ptrs := dest

			if len(dest) == 1 {
				if inner, ok := dest[0].([]any); ok {
					ptrs = inner
				}
			}

			fillEventDest(ptrs, event)

			if stateBeforePtr, ok := ptrs[9].(*[]byte); ok {
				*stateBeforePtr = []byte(`{invalid json`)
			}
		}).
		Return(nil).
		Once()
}

func scanCorruptStateAfterIntoRow(row *MockRow, event domain.AuditEvent) {
	row.EXPECT().
		Scan(mock.Anything).
		Run(func(dest ...any) {
			ptrs := dest

			if len(dest) == 1 {
				if inner, ok := dest[0].([]any); ok {
					ptrs = inner
				}
			}

			fillEventDest(ptrs, event)

			if stateAfterPtr, ok := ptrs[10].(*[]byte); ok {
				*stateAfterPtr = []byte(`{invalid json`)
			}
		}).
		Return(nil).
		Once()
}
