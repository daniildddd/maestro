//go:build integration

package audit_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	core_logger "github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/repository/postgres/pgxadapter"
	"github.com/daniildddd/maestro/internal/features/audit/repository"
)

const (
	testDBUser     = "maestro"
	testDBPassword = "maestro"
	testDBName     = "maestro"
)

func newAuditRepository(t *testing.T) *repository.AuditRepository {
	t.Helper()

	ctx := context.Background()
	must := require.New(t)

	container, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase(testDBName),
		postgres.WithUsername(testDBUser),
		postgres.WithPassword(testDBPassword),
		postgres.WithOrderedInitScripts(
			"../../../migrations/000001_create_users_table.up.sql",
			"../../../migrations/000002_create_refresh_tokens_table.up.sql",
			"../../../migrations/000003_create_audit_logs_table.up.sql",
			"../../../migrations/000004_create_refresh_tokens_expires_at_index.up.sql",
			"../../../migrations/000005_create_audit_logs_table.up.sql",
		),
		testcontainers.WithWaitStrategy(
			wait.ForSQL("5432/tcp", "pgx", func(host string, port network.Port) string {
				return fmt.Sprintf(
					"postgres://%s:%s@%s/%s?sslmode=disable",
					testDBUser,
					testDBPassword,
					net.JoinHostPort(host, port.Port()),
					testDBName,
				)
			}).
				WithQuery("SELECT 1 FROM audit_logs LIMIT 1").
				WithStartupTimeout(60*time.Second),
		),
	)
	must.NoError(err)

	testcontainers.CleanupContainer(t, container)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	must.NoError(err)

	pool, err := pgxadapter.NewPool(ctx, pgxadapter.Config{DSN: dsn, Timeout: 5 * time.Second})
	must.NoError(err)

	t.Cleanup(pool.Close)

	return repository.NewAuditRepository(pool)
}

func recordContext() context.Context {
	return core_logger.ToContext(context.Background(), &core_logger.Logger{Logger: zap.NewNop()})
}

func mustEvent(action domain.Action, actor string, createdAt time.Time) domain.AuditEvent {
	return domain.AuditEvent{
		ID:          uuid.New(),
		Action:      action,
		Outcome:     domain.OutcomeSuccess,
		ActorLogin:  actor,
		SubjectType: domain.AuditSubjectUser,
		SubjectID:   "subject-" + actor,
		SubjectName: actor,
		StateAfter:  map[string]any{"role": "user"},
		RequestID:   "req-" + actor,
		IP:          "203.0.113.7",
		UserAgent:   "integration-test",
		CreatedAt:   createdAt,
	}
}

func assertEventRoundTrip(t *testing.T, want, got domain.AuditEvent) {
	t.Helper()

	is := assert.New(t)
	is.Equal(want.ID, got.ID)
	is.Equal(want.Action, got.Action)
	is.Equal(want.Outcome, got.Outcome)
	is.Equal(want.FailureReason, got.FailureReason)
	is.Equal(want.ActorID, got.ActorID)
	is.Equal(want.ActorLogin, got.ActorLogin)
	is.Equal(want.SubjectType, got.SubjectType)
	is.Equal(want.SubjectID, got.SubjectID)
	is.Equal(want.SubjectName, got.SubjectName)
	is.Equal(want.StateAfter, got.StateAfter)
	is.True(want.StateBefore == nil && got.StateBefore == nil, "nil state should survive round-trip")
	is.Equal(want.RequestID, got.RequestID)
	is.Equal(want.IP, got.IP)
	is.Equal(want.UserAgent, got.UserAgent)
	is.True(want.CreatedAt.Equal(got.CreatedAt), "created_at should survive round-trip")
}

func TestAuditRepository_RecordAndGetLog(t *testing.T) {
	t.Parallel()

	must := require.New(t)
	repo := newAuditRepository(t)
	ctx := recordContext()

	recorded := mustEvent(domain.ActionAuthLogin, "alice", time.Now().UTC().Truncate(time.Microsecond))
	repo.Record(ctx, recorded)

	got, err := repo.GetLogByID(context.Background(), recorded.ID)
	must.NoError(err)

	assertEventRoundTrip(t, recorded, got)

	_, err = repo.GetLogByID(context.Background(), uuid.New())
	must.ErrorIs(err, errs.ErrAuditLogNotFound)
}

func TestAuditRepository_GetLogsFiltering(t *testing.T) {
	t.Parallel()

	repo := newAuditRepository(t)
	ctx := recordContext()

	base := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	seeded := []domain.AuditEvent{
		mustEvent(domain.ActionAuthLogin, "alice", base.Add(time.Second)),
		mustEvent(domain.ActionAuthLogin, "bob", base.Add(2*time.Second)),
		mustEvent(domain.ActionUserCreated, "alice", base.Add(3*time.Second)),
	}

	for _, event := range seeded {
		repo.Record(ctx, event)
	}

	tests := []struct {
		name   string
		page   int
		limit  int
		action string
		actor  string
		want   []domain.AuditEvent
	}{
		{
			name:  "first page",
			page:  1,
			limit: 1,
			want:  []domain.AuditEvent{seeded[2], seeded[1]},
		},
		{
			name:  "second page",
			page:  2,
			limit: 1,
			want:  []domain.AuditEvent{seeded[1], seeded[0]},
		},
		{
			name:   "filter by action",
			page:   1,
			limit:  20,
			action: string(domain.ActionAuthLogin),
			want:   []domain.AuditEvent{seeded[1], seeded[0]},
		},
		{
			name:  "filter by actor",
			page:  1,
			limit: 20,
			actor: "alice",
			want:  []domain.AuditEvent{seeded[2], seeded[0]},
		},
		{
			name:   "filter by action and actor",
			page:   1,
			limit:  20,
			action: string(domain.ActionAuthLogin),
			actor:  "alice",
			want:   []domain.AuditEvent{seeded[0]},
		},
		{
			name:   "filter by action and actor mismatch returns empty",
			page:   1,
			limit:  20,
			action: string(domain.ActionUserCreated),
			actor:  "bob",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			filter, err := domain.NewAuditLogFilter(tt.page, tt.limit, tt.action, tt.actor)
			must.NoError(err)

			logs, err := repo.GetLogs(context.Background(), filter)
			must.NoError(err)
			must.Len(logs, len(tt.want))

			for i := range tt.want {
				assertEventRoundTrip(t, tt.want[i], logs[i])
			}
		})
	}
}

func TestAuditRepository_DeleteLogByID(t *testing.T) {
	t.Parallel()

	must := require.New(t)
	repo := newAuditRepository(t)
	ctx := recordContext()

	recorded := mustEvent(domain.ActionAuthLogout, "alice", time.Now().UTC().Truncate(time.Microsecond))
	repo.Record(ctx, recorded)
	must.NoError(repo.DeleteLogByID(context.Background(), recorded.ID))

	_, err := repo.GetLogByID(context.Background(), recorded.ID)
	must.ErrorIs(err, errs.ErrAuditLogNotFound)

	must.ErrorIs(repo.DeleteLogByID(context.Background(), uuid.New()), errs.ErrAuditLogNotFound)
}
