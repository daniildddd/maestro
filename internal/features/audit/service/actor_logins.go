package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func resolveActorLogins(
	ctx context.Context,
	users UserDirectory,
	events []domain.AuditEvent,
) error {
	const op = "audit.service.resolveActorLogins"

	ids := make([]uuid.UUID, 0, len(events))
	seen := make(map[uuid.UUID]struct{})

	for _, e := range events {
		if e.ActorLogin != "" || e.ActorID == uuid.Nil {
			continue
		}

		if _, ok := seen[e.ActorID]; !ok {
			seen[e.ActorID] = struct{}{}
			ids = append(ids, e.ActorID)
		}
	}

	if len(ids) == 0 {
		return nil
	}

	found, err := users.GetUsersByIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	byID := make(map[uuid.UUID]string, len(found))

	for _, u := range found {
		byID[u.ID] = u.Username
	}

	for i := range events {
		if events[i].ActorLogin == "" {
			events[i].ActorLogin = byID[events[i].ActorID]
		}
	}

	return nil
}
