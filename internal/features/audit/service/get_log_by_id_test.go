package service_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/features/audit/service"
)

func TestAuditService_GetLogByID(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	event := mustNewAuditEvent(id)

	tests := []struct {
		name      string
		setupMock func(m *MockAuditRepository)
		wantBody  domain.AuditEvent
		wantIs    error
	}{
		{
			name: "success returns event",
			setupMock: func(m *MockAuditRepository) {
				m.EXPECT().
					GetLogByID(mock.Anything, id).
					Return(event, nil).
					Once()
			},
			wantBody: event,
		},
		{
			name: "repository error is wrapped",
			setupMock: func(m *MockAuditRepository) {
				m.EXPECT().
					GetLogByID(mock.Anything, id).
					Return(domain.AuditEvent{}, errs.ErrAuditLogNotFound).
					Once()
			},
			wantIs: errs.ErrAuditLogNotFound,
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

			got, err := svc.GetLogByID(t.Context(), id)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			is.Equal(tt.wantBody, got)
		})
	}
}
