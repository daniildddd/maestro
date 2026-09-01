package transport

import (
	"time"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
)

type AuditEventDTOResponse struct {
	ID            string         `json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	Action        string         `json:"action"`
	Outcome       string         `json:"outcome"`
	FailureReason *string        `json:"failure_reason"`
	Actor         *ActorDTO      `json:"actor"`
	Subject       SubjectDTO     `json:"subject"`
	StateBefore   map[string]any `json:"state_before"`
	StateAfter    map[string]any `json:"state_after"`
	Request       *RequestDTO    `json:"request"`
}

type ActorDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type SubjectDTO struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RequestDTO struct {
	ID        string `json:"id"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
}

func auditEventDTOFromDomain(event domain.AuditEvent) AuditEventDTOResponse {
	var actor *ActorDTO

	if event.ActorID != uuid.Nil {
		actor = &ActorDTO{
			ID:       event.ActorID.String(),
			Username: event.ActorLogin,
		}
	}

	var failureReason *string

	if event.FailureReason != "" {
		failureReason = &event.FailureReason
	}

	var request *RequestDTO

	if event.RequestID != "" {
		request = &RequestDTO{
			ID:        event.RequestID,
			IP:        event.IP,
			UserAgent: event.UserAgent,
		}
	}

	return AuditEventDTOResponse{
		ID:            event.ID.String(),
		CreatedAt:     event.CreatedAt,
		Action:        string(event.Action),
		Outcome:       string(event.Outcome),
		FailureReason: failureReason,
		Actor:         actor,
		Subject: SubjectDTO{
			Type: event.SubjectType,
			ID:   event.SubjectID,
			Name: event.SubjectName,
		},
		StateBefore: event.StateBefore,
		StateAfter:  event.StateAfter,
		Request:     request,
	}
}

func auditEventsDTOFromDomains(events []domain.AuditEvent) []AuditEventDTOResponse {
	dtos := make([]AuditEventDTOResponse, 0, len(events))

	for _, event := range events {
		dtos = append(dtos, auditEventDTOFromDomain(event))
	}

	return dtos
}
