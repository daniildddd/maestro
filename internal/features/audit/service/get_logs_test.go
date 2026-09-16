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

	sharedID := uuid.New()
	missingID := uuid.New()

	emptyOne := mustNewAuditEvent(uuid.New())
	emptyOne.ActorID = sharedID
	emptyOne.ActorLogin = ""
	emptyTwo := mustNewAuditEvent(uuid.New())
	emptyTwo.ActorID = sharedID
	emptyTwo.ActorLogin = ""
	filled := mustNewAuditEvent(uuid.New())
	missing := mustNewAuditEvent(uuid.New())
	missing.ActorID = missingID
	missing.ActorLogin = ""

	filledOne := emptyOne
	filledOne.ActorLogin = "bob"
	filledTwo := emptyTwo
	filledTwo.ActorLogin = "bob"

	tests := []struct {
		name       string
		filter     *domain.AuditLogFilter
		setupMock  func(m *MockAuditRepository)
		setupUsers func(m *MockUserDirectory)
		wantBody   []domain.AuditEvent
		wantMore   bool
		wantIs     error
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
		{
			name:   "empty actor logins are enriched from users",
			filter: &domain.AuditLogFilter{Page: 1, Limit: 20},
			setupMock: func(m *MockAuditRepository) {
				m.EXPECT().
					GetLogs(mock.Anything, mock.Anything).
					Return([]domain.AuditEvent{emptyOne, emptyTwo, filled, missing}, nil).
					Once()
			},
			setupUsers: func(m *MockUserDirectory) {
				m.EXPECT().
					GetUsersByIDs(mock.Anything, mock.MatchedBy(func(ids []uuid.UUID) bool {
						return len(ids) == 2 && ids[0] == sharedID && ids[1] == missingID
					})).
					Return([]domain.User{{ID: sharedID, Username: "bob"}}, nil).
					Once()
			},
			wantBody: []domain.AuditEvent{filledOne, filledTwo, filled, missing},
			wantMore: false,
		},
		{
			name:   "users lookup error is wrapped",
			filter: &domain.AuditLogFilter{Page: 1, Limit: 20},
			setupMock: func(m *MockAuditRepository) {
				m.EXPECT().
					GetLogs(mock.Anything, mock.Anything).
					Return([]domain.AuditEvent{emptyOne}, nil).
					Once()
			},
			setupUsers: func(m *MockUserDirectory) {
				m.EXPECT().
					GetUsersByIDs(mock.Anything, mock.Anything).
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

			users := NewMockUserDirectory(t)

			if tt.setupUsers != nil {
				tt.setupUsers(users)
			}

			svc := service.NewAuditService(repo, users)

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
