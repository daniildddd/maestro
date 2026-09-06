package repository_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	core_logger "github.com/daniildddd/maestro/internal/core/logger"
)

func TestRecord(t *testing.T) {
	t.Parallel()

	const opTimeout = 100 * time.Millisecond

	tests := []struct {
		name    string
		event   domain.AuditEvent
		setup   func(pool *MockPool)
		wantLog bool
	}{
		{
			name: "full event is inserted with all columns",
			event: domain.AuditEvent{
				ID:          uuid.New(),
				Action:      domain.ActionAuthLogin,
				Outcome:     domain.OutcomeSuccess,
				ActorID:     uuid.New(),
				ActorLogin:  "alice",
				SubjectType: domain.AuditSubjectUser,
				SubjectID:   "sub-id",
				SubjectName: "alice",
				RequestID:   "req-1",
				IP:          "203.0.113.7",
				UserAgent:   "test-agent",
				CreatedAt:   time.Now().UTC(),
			},
			setup: func(pool *MockPool) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Exec(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, nil).
					Once()
			},
		},
		{
			name: "minimal event with nil states is inserted",
			event: domain.AuditEvent{
				ID:          uuid.New(),
				Action:      domain.ActionUserDeleted,
				Outcome:     domain.OutcomeSuccess,
				SubjectType: domain.AuditSubjectUser,
				CreatedAt:   time.Now().UTC(),
			},
			setup: func(pool *MockPool) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Exec(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, nil).
					Once()
			},
		},
		{
			name: "exec error is logged and does not propagate",
			event: domain.AuditEvent{
				ID:          uuid.New(),
				Action:      domain.ActionConnectorCreated,
				Outcome:     domain.OutcomeSuccess,
				SubjectType: domain.AuditSubjectConnector,
				CreatedAt:   time.Now().UTC(),
			},
			setup: func(pool *MockPool) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Exec(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, assert.AnError).
					Once()
			},
			wantLog: true,
		},
		{
			name: "unserializable state before is logged",
			event: domain.AuditEvent{
				ID:          uuid.New(),
				Action:      domain.ActionAuthLogin,
				Outcome:     domain.OutcomeSuccess,
				SubjectType: domain.AuditSubjectUser,
				StateBefore: map[string]any{"bad": make(chan int)},
				CreatedAt:   time.Now().UTC(),
			},
			setup: func(pool *MockPool) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()
			},
			wantLog: true,
		},
		{
			name: "unserializable state after is logged",
			event: domain.AuditEvent{
				ID:          uuid.New(),
				Action:      domain.ActionAuthLogin,
				Outcome:     domain.OutcomeSuccess,
				SubjectType: domain.AuditSubjectUser,
				StateAfter:  map[string]any{"bad": make(chan int)},
				CreatedAt:   time.Now().UTC(),
			},
			setup: func(pool *MockPool) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()
			},
			wantLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			must := require.New(t)

			pool := NewMockPool(t)
			tt.setup(pool)

			repo := newAuditRepo(pool)

			sink, logged := observer.New(zapcore.ErrorLevel)
			observed := &core_logger.Logger{Logger: zap.New(sink)}

			must.NotPanics(func() {
				repo.Record(core_logger.ToContext(context.Background(), observed), tt.event)
			})

			entries := logged.All()

			if !tt.wantLog {
				is.Empty(entries)

				return
			}

			must.Len(entries, 1)
			is.Equal("record audit event", entries[0].Message)

			fields := entries[0].ContextMap()
			is.Equal(string(tt.event.Action), fields["action"])
			is.Equal(string(tt.event.Outcome), fields["outcome"])

			_, ok := fields["error"]
			must.True(ok, "record audit event log must carry error field")
		})
	}
}
