package service_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/features/audit/service"
)

func TestAuditService_DeleteLogByID(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	tests := []struct {
		name      string
		setupMock func(m *MockAuditRepository)
		wantIs    error
	}{
		{
			name: "success deletes log",
			setupMock: func(m *MockAuditRepository) {
				m.EXPECT().
					DeleteLogByID(mock.Anything, id).
					Return(nil).
					Once()
			},
		},
		{
			name: "repository error is wrapped",
			setupMock: func(m *MockAuditRepository) {
				m.EXPECT().
					DeleteLogByID(mock.Anything, id).
					Return(errs.ErrAuditLogNotFound).
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

			err := svc.DeleteLogByID(t.Context(), id)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
		})
	}
}
