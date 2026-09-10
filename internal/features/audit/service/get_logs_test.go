package service_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/features/audit/service"
)

func mustNewAuditEvent(id uuid.UUID) domain.AuditEvent {
	return domain.AuditEvent{
		ID:          id,
		Action:      domain.ActionAuthLogin,
		Outcome:     domain.OutcomeSuccess,
		ActorID:     uuid.New(),
		ActorLogin:  "alice",
		SubjectType: domain.AuditSubjectUser,
		SubjectID:   id.String(),
		SubjectName: "alice",
		CreatedAt:   time.Now().UTC(),
	}
}

func TestAuditService_GetLogs(t *testing.T) {
	t.Parallel()

	events := []domain.AuditEvent{
		mustNewAuditEvent(uuid.New()),
		mustNewAuditEvent(uuid.New()),
		mustNewAuditEvent(uuid.New()),
	}

	tests := []struct {
		name      string
		filter    *domain.AuditLogFilter
		setupMock func(m *MockAuditRepository)
		wantBody  []domain.AuditEvent
		wantMore  bool
		wantIs    error
	}{
		{
			name:   "success returns events without trim",
			filter: &domain.AuditLogFilter{Page: 1, Limit: 3},
			setupMock: func(m *MockAuditRepository) {
				m.EXPECT().
					GetLogs(mock.Anything, mock.Anything).
					Return(events, nil).
					Once()
			},
			wantBody: events,
			wantMore: false,
		},
		{
			name:   "events above limit are trimmed and has more set",
			filter: &domain.AuditLogFilter{Page: 1, Limit: 2},
			setupMock: func(m *MockAuditRepository) {
				m.EXPECT().
					GetLogs(mock.Anything, mock.Anything).
					Return(events, nil).
					Once()
			},
			wantBody: events[:2],
			wantMore: true,
		},
		{
			name:   "repository error is wrapped",
			filter: &domain.AuditLogFilter{Page: 1, Limit: 20},
			setupMock: func(m *MockAuditRepository) {
				m.EXPECT().
					GetLogs(mock.Anything, mock.Anything).
					Return(nil, errs.ErrInternal).
					Once()
			},
			wantIs: errs.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			must := require.New(t)

			repo := NewMockAuditRepository(t)
			tt.setupMock(repo)

			svc := service.NewAuditService(repo)

			got, hasMore, err := svc.GetLogs(t.Context(), tt.filter)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			is.Equal(tt.wantMore, hasMore)
			is.Equal(tt.wantBody, got)
		})
	}
}
