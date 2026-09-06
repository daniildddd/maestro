package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
)

func TestGetLogs(t *testing.T) {
	t.Parallel()

	const opTimeout = 100 * time.Millisecond

	id := uuid.New()

	const queryNoFilters = "\n\tSELECT id, action, outcome, failure_reason, actor_id, actor_login,\n\t       subject_type, subject_id, subject_name, state_before, state_after,\n\t       request_id, ip, user_agent, created_at\n\tFROM audit_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2"

	tests := []struct {
		name       string
		filter     *domain.AuditLogFilter
		setup      func(pool *MockPool, rows *MockRows)
		wantErr    bool
		wantIs     error
		wantEvents []domain.AuditEvent
	}{
		{
			name:   "success scans events without filters",
			filter: &domain.AuditLogFilter{Page: 1, Limit: 20},
			setup: func(pool *MockPool, rows *MockRows) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Query(
						mock.Anything,
						queryNoFilters,
						[]any{21, 0},
					).
					Return(rows, nil).
					Once()

				rows.EXPECT().
					Next().
					Return(true).
					Once()

				scanEventIntoRows(rows, mustAuditRow(id))

				rows.EXPECT().
					Next().
					Return(false).
					Once()

				rows.EXPECT().
					Err().
					Return(nil).
					Once()

				rows.EXPECT().
					Close().
					Once()
			},
			wantEvents: []domain.AuditEvent{
				mustAuditEvent(t, id),
			},
		},
		{
			name:   "action and actor filters produce where clause",
			filter: &domain.AuditLogFilter{Page: 2, Limit: 10, Action: domain.ActionAuthLogin, Actor: "alice"},
			setup: func(pool *MockPool, rows *MockRows) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Query(
						mock.Anything,
						"\n\tSELECT id, action, outcome, failure_reason, actor_id, actor_login,\n\t       subject_type, subject_id, subject_name, state_before, state_after,\n\t       request_id, ip, user_agent, created_at\n\tFROM audit_logs WHERE action=$1 AND actor_login=$2 ORDER BY created_at DESC LIMIT $3 OFFSET $4",
						[]any{"auth.login", "alice", 11, 10},
					).
					Return(rows, nil).
					Once()

				rows.EXPECT().
					Next().
					Return(false).
					Once()

				rows.EXPECT().
					Err().
					Return(nil).
					Once()

				rows.EXPECT().
					Close().
					Once()
			},
			wantEvents: nil,
		},
		{
			name:   "query error is wrapped",
			filter: &domain.AuditLogFilter{Page: 1, Limit: 20},
			setup: func(pool *MockPool, _ *MockRows) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Query(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errs.ErrInternal).
					Once()
			},
			wantErr: true,
			wantIs:  errs.ErrInternal,
		},
		{
			name:   "scan error is wrapped",
			filter: &domain.AuditLogFilter{Page: 1, Limit: 20},
			setup: func(pool *MockPool, rows *MockRows) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Query(mock.Anything, mock.Anything, mock.Anything).
					Return(rows, nil).
					Once()

				rows.EXPECT().
					Next().
					Return(true).
					Once()

				rows.EXPECT().
					Scan(mock.Anything).
					Return(errs.ErrInternal).
					Once()

				rows.EXPECT().
					Close().
					Once()
			},
			wantErr: true,
			wantIs:  errs.ErrInternal,
		},
		{
			name:   "rows error is wrapped",
			filter: &domain.AuditLogFilter{Page: 1, Limit: 20},
			setup: func(pool *MockPool, rows *MockRows) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Query(mock.Anything, mock.Anything, mock.Anything).
					Return(rows, nil).
					Once()

				rows.EXPECT().
					Next().
					Return(false).
					Once()

				rows.EXPECT().
					Err().
					Return(errs.ErrInternal).
					Once()

				rows.EXPECT().
					Close().
					Once()
			},
			wantErr: true,
			wantIs:  errs.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			must := require.New(t)

			pool := NewMockPool(t)
			rows := NewMockRows(t)
			tt.setup(pool, rows)

			repo := newAuditRepo(pool)

			events, err := repo.GetLogs(context.Background(), tt.filter)

			if tt.wantErr {
				must.Error(err)

				if tt.wantIs != nil {
					must.ErrorIs(err, tt.wantIs)
				}

				return
			}

			must.NoError(err)
			is.Equal(tt.wantEvents, events)
		})
	}
}
