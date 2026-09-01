package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
)

type auditEventModel struct {
	ID            uuid.UUID
	Action        string
	Outcome       string
	FailureReason *string
	ActorID       *uuid.UUID
	ActorLogin    *string
	SubjectType   string
	SubjectID     *string
	SubjectName   *string
	StateBefore   []byte
	StateAfter    []byte
	RequestID     *string
	IP            *string
	UserAgent     *string
	CreatedAt     time.Time
}

func (m *auditEventModel) Scan(row core_postgres_pool.Row) error {
	return row.Scan(
		&m.ID,
		&m.Action,
		&m.Outcome,
		&m.FailureReason,
		&m.ActorID,
		&m.ActorLogin,
		&m.SubjectType,
		&m.SubjectID,
		&m.SubjectName,
		&m.StateBefore,
		&m.StateAfter,
		&m.RequestID,
		&m.IP,
		&m.UserAgent,
		&m.CreatedAt,
	)
}

func (m *auditEventModel) toDomain() (domain.AuditEvent, error) {
	const op = "audit.repository.toDomain"

	event := domain.AuditEvent{
		ID:          m.ID,
		Action:      domain.Action(m.Action),
		Outcome:     domain.Outcome(m.Outcome),
		SubjectType: m.SubjectType,
		CreatedAt:   m.CreatedAt,
	}

	if m.FailureReason != nil {
		event.FailureReason = *m.FailureReason
	}

	if m.ActorID != nil {
		event.ActorID = *m.ActorID
	}

	if m.ActorLogin != nil {
		event.ActorLogin = *m.ActorLogin
	}

	if m.SubjectID != nil {
		event.SubjectID = *m.SubjectID
	}

	if m.SubjectName != nil {
		event.SubjectName = *m.SubjectName
	}

	if m.StateBefore != nil {
		if err := json.Unmarshal(m.StateBefore, &event.StateBefore); err != nil {
			return domain.AuditEvent{}, fmt.Errorf("%s: unmarshal state_before: %w", op, err)
		}
	}

	if m.StateAfter != nil {
		if err := json.Unmarshal(m.StateAfter, &event.StateAfter); err != nil {
			return domain.AuditEvent{}, fmt.Errorf("%s: unmarshal state_after: %w", op, err)
		}
	}

	if m.RequestID != nil {
		event.RequestID = *m.RequestID
	}

	if m.IP != nil {
		event.IP = *m.IP
	}

	if m.UserAgent != nil {
		event.UserAgent = *m.UserAgent
	}

	return event, nil
}
