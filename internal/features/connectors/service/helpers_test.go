package service_test

import (
	"context"

	"github.com/daniildddd/maestro/internal/core/domain"
)

type noopAuditor struct{}

func (noopAuditor) Record(context.Context, domain.AuditEvent) {
}
