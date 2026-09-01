package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Action string

const (
	ActionAuthLogin   Action = "auth.login"
	ActionAuthRefresh Action = "auth.refresh"
	ActionAuthLogout  Action = "auth.logout"

	ActionUserCreated         Action = "user.created"
	ActionUserUpdated         Action = "user.updated"
	ActionUserDeleted         Action = "user.deleted"
	ActionUserPasswordChanged Action = "user.password_changed"
	ActionUserPasswordReset   Action = "user.password_reset"

	ActionConnectorCreated   Action = "connector.created"
	ActionConnectorUpdated   Action = "connector.updated"
	ActionConnectorDeleted   Action = "connector.deleted"
	ActionConnectorPaused    Action = "connector.paused"
	ActionConnectorResumed   Action = "connector.resumed"
	ActionConnectorRestarted Action = "connector.restarted"
)

type Outcome string

const (
	OutcomeSuccess Outcome = "success"
	OutcomeFailure Outcome = "failure"
)

const (
	AuditSubjectUser      = "user"
	AuditSubjectConnector = "connector"
)

var ErrInvalidAuditAction = errors.New("action must be one of the known audit actions")

type AuditEvent struct {
	ID            uuid.UUID
	Action        Action
	Outcome       Outcome
	FailureReason string
	ActorID       uuid.UUID
	ActorLogin    string
	SubjectType   string
	SubjectID     string
	SubjectName   string
	StateBefore   map[string]any
	StateAfter    map[string]any
	RequestID     string
	IP            string
	UserAgent     string
	CreatedAt     time.Time
}

type AuditLogFilter struct {
	Page   int
	Limit  int
	Action Action
	Actor  string
}

func NewAuditLogFilter(
	page int,
	limit int,
	action string,
	actor string,
) (*AuditLogFilter, error) {
	const op = "core.domain.NewAuditLogFilter"

	f := &AuditLogFilter{
		Page:   page,
		Limit:  limit,
		Action: Action(action),
		Actor:  actor,
	}
	f.Normalize()

	if err := f.Validate(); err != nil {
		return nil, fmt.Errorf(
			"%s: %w", op, err,
		)
	}

	return f, nil
}

func (f *AuditLogFilter) Normalize() {
	if f.Page <= 0 {
		f.Page = 1
	}

	if f.Limit <= 0 {
		f.Limit = 20
	}

	if f.Limit > 100 {
		f.Limit = 100
	}
}

func (f *AuditLogFilter) Validate() error {
	const op = "domain.AuditLogFilter.Validate"

	if f.Action != "" && !f.Action.IsValid() {
		return fmt.Errorf("%s: action=%s: %w", op, f.Action, ErrInvalidAuditAction)
	}

	return nil
}

func (a Action) IsValid() bool {
	switch a {
	case ActionAuthLogin,
		ActionAuthRefresh,
		ActionAuthLogout,
		ActionUserCreated,
		ActionUserUpdated,
		ActionUserDeleted,
		ActionUserPasswordChanged,
		ActionUserPasswordReset,
		ActionConnectorCreated,
		ActionConnectorUpdated,
		ActionConnectorDeleted,
		ActionConnectorPaused,
		ActionConnectorResumed,
		ActionConnectorRestarted:
		return true
	default:
		return false
	}
}
