//go:build integration

package users_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/google/uuid"
	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/repository/postgres/pgxadapter"
	"github.com/daniildddd/maestro/internal/features/users/repository"
)

const (
	testDBUser     = "maestro"
	testDBPassword = "maestro"
	testDBName     = "maestro"
)

func newUsersRepository(t *testing.T) *repository.UsersRepository {
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
				WithQuery("SELECT 1 FROM users LIMIT 1").
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

	return repository.NewUsersRepository(pool)
}

func mustUser(t *testing.T, username, role string, createdAt time.Time) domain.User {
	t.Helper()

	user, err := domain.NewUser(uuid.New(), username, "hash-for-"+username, role, createdAt, nil)
	require.NoError(t, err)

	return user
}

func assertUserRoundTrip(t *testing.T, want, got domain.User) {
	t.Helper()

	is := assert.New(t)
	is.Equal(want.ID, got.ID)
	is.Equal(want.Username, got.Username)
	is.Equal(want.PasswordHash, got.PasswordHash)
	is.Equal(want.Role, got.Role)
	is.True(want.CreatedAt.Equal(got.CreatedAt), "created_at should survive round-trip")
	is.Nil(got.UpdatedAt)
}

func TestUsersRepository_CreateAndGetUser(t *testing.T) {
	t.Parallel()

	must := require.New(t)
	repo := newUsersRepository(t)
	ctx := context.Background()

	want := mustUser(t, "alice", domain.RoleUser, time.Now().UTC().Truncate(time.Microsecond))

	created, err := repo.CreateUser(ctx, want)
	must.NoError(err)

	assertUserRoundTrip(t, want, created)

	got, err := repo.GetUserByID(ctx, want.ID)
	must.NoError(err)

	assertUserRoundTrip(t, want, got)

	_, err = repo.GetUserByID(ctx, uuid.New())
	must.ErrorIs(err, errs.ErrUserNotFound)
}

func TestUsersRepository_CreateUserConflict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		conflict func(u domain.User) domain.User
		wantIs   error
	}{
		{
			name: "duplicate username maps to username conflict",
			conflict: func(u domain.User) domain.User {
				u.ID = uuid.New()

				return u
			},
			wantIs: errs.ErrUsernameConflict,
		},
		{
			name: "duplicate id is reported as username conflict",
			conflict: func(u domain.User) domain.User {
				u.Username = "robert"

				return u
			},
			wantIs: errs.ErrUsernameConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)
			repo := newUsersRepository(t)
			ctx := context.Background()

			original := mustUser(t, "bob", domain.RoleUser, time.Now().UTC().Truncate(time.Microsecond))

			_, err := repo.CreateUser(ctx, original)
			must.NoError(err)

			_, err = repo.CreateUser(ctx, tt.conflict(original))
			must.ErrorIs(err, tt.wantIs)
		})
	}
}

func TestUsersRepository_GetUsers(t *testing.T) {
	t.Parallel()

	must := require.New(t)
	repo := newUsersRepository(t)
	ctx := context.Background()

	base := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	seeded := []domain.User{
		mustUser(t, "user1", domain.RoleUser, base.Add(time.Second)),
		mustUser(t, "user2", domain.RoleUser, base.Add(2*time.Second)),
		mustUser(t, "user3", domain.RoleAdmin, base.Add(3*time.Second)),
	}

	for _, u := range seeded {
		_, err := repo.CreateUser(ctx, u)
		must.NoError(err)
	}

	tests := []struct {
		name     string
		page     int
		limit    int
		username string
		role     string
		want     []domain.User
	}{
		{
			name:  "first page",
			page:  1,
			limit: 2,
			want:  []domain.User{seeded[0], seeded[1], seeded[2]},
		},
		{
			name:  "second page",
			page:  2,
			limit: 2,
			want:  []domain.User{seeded[2]},
		},
		{
			name:     "filter by username",
			page:     1,
			limit:    20,
			username: "user2",
			want:     []domain.User{seeded[1]},
		},
		{
			name:  "filter by role",
			page:  1,
			limit: 20,
			role:  domain.RoleAdmin,
			want:  []domain.User{seeded[2]},
		},
		{
			name:     "filter by username and role",
			page:     1,
			limit:    20,
			username: "user3",
			role:     domain.RoleAdmin,
			want:     []domain.User{seeded[2]},
		},
		{
			name:     "filter by username and role mismatch returns empty",
			page:     1,
			limit:    20,
			username: "user2",
			role:     domain.RoleAdmin,
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			filter, err := domain.NewUserFilter(tt.page, tt.limit, tt.username, tt.role)
			must.NoError(err)

			users, err := repo.GetUsers(ctx, filter)
			must.NoError(err)
			must.Len(users, len(tt.want))

			for i := range tt.want {
				assertUserRoundTrip(t, tt.want[i], users[i])
			}
		})
	}
}
