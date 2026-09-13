//go:build integration

package auth_test

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
	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
	"github.com/daniildddd/maestro/internal/core/repository/postgres/pgxadapter"
	"github.com/daniildddd/maestro/internal/features/auth/repository"
)

const (
	testDBUser     = "maestro"
	testDBPassword = "maestro"
	testDBName     = "maestro"
)

func newAuthRepository(t *testing.T) (*repository.AuthRepository, core_postgres_pool.Pool) {
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
				WithQuery("SELECT 1 FROM refresh_tokens LIMIT 1").
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

	return repository.NewAuthRepository(pool), pool
}

func seedUser(t *testing.T, pool core_postgres_pool.Pool) uuid.UUID {
	t.Helper()

	must := require.New(t)
	id := uuid.New()

	_, err := pool.Exec(context.Background(),
		`INSERT INTO users (id, username, password_hash, role, created_at)
		 VALUES ($1, $2, $3, $4, now())`,
		id, "owner-"+id.String()[:8], "hash", domain.RoleUser,
	)
	must.NoError(err)

	return id
}

func mustToken(userID uuid.UUID, hash string, expiresAt time.Time) domain.RefreshToken {
	return domain.NewRefreshToken(
		uuid.New(),
		userID,
		hash,
		time.Now().UTC().Truncate(time.Microsecond),
		expiresAt.Truncate(time.Microsecond),
	)
}

func assertTokenRoundTrip(t *testing.T, want, got domain.RefreshToken) {
	t.Helper()

	is := assert.New(t)
	is.Equal(want.ID, got.ID)
	is.Equal(want.UserID, got.UserID)
	is.Equal(want.TokenHash, got.TokenHash)
	is.True(want.CreatedAt.Equal(got.CreatedAt), "created_at should survive round-trip")
	is.True(want.ExpiresAt.Equal(got.ExpiresAt), "expires_at should survive round-trip")
}

func TestAuthRepository_SaveAndGetRefreshToken(t *testing.T) {
	t.Parallel()

	must := require.New(t)
	repo, pool := newAuthRepository(t)
	ctx := context.Background()

	userID := seedUser(t, pool)
	saved := mustToken(userID, "hash-save-get", time.Now().UTC().Add(24*time.Hour))

	must.NoError(repo.SaveRefreshToken(ctx, saved))

	got, err := repo.GetRefreshTokenByHash(ctx, saved.TokenHash)
	must.NoError(err)

	assertTokenRoundTrip(t, saved, got)

	_, err = repo.GetRefreshTokenByHash(ctx, "hash-missing")
	must.ErrorIs(err, errs.ErrRefreshTokenNotFound)
}

func TestAuthRepository_SaveRefreshTokenUnknownUser(t *testing.T) {
	t.Parallel()

	must := require.New(t)
	repo, _ := newAuthRepository(t)
	ctx := context.Background()

	err := repo.SaveRefreshToken(ctx, mustToken(
		uuid.New(),
		"hash-orphan",
		time.Now().UTC().Add(time.Hour),
	))
	must.ErrorIs(err, core_postgres_pool.ErrViolatesForeignKey)
}

func TestAuthRepository_RotateRefreshToken(t *testing.T) {
	t.Parallel()

	must := require.New(t)
	repo, pool := newAuthRepository(t)
	ctx := context.Background()

	userID := seedUser(t, pool)
	oldToken := mustToken(userID, "hash-old", time.Now().UTC().Add(time.Hour))
	must.NoError(repo.SaveRefreshToken(ctx, oldToken))

	newToken := mustToken(userID, "hash-new", time.Now().UTC().Add(2*time.Hour))
	must.NoError(repo.RotateRefreshToken(ctx, oldToken.TokenHash, newToken))

	_, err := repo.GetRefreshTokenByHash(ctx, oldToken.TokenHash)
	must.ErrorIs(err, errs.ErrRefreshTokenNotFound)

	got, err := repo.GetRefreshTokenByHash(ctx, newToken.TokenHash)
	must.NoError(err)

	assertTokenRoundTrip(t, newToken, got)

	must.ErrorIs(
		repo.RotateRefreshToken(
			ctx,
			"hash-missing",
			mustToken(userID, "hash-other", time.Now().UTC().Add(time.Hour)),
		),
		errs.ErrRefreshTokenNotFound,
	)
}

func TestAuthRepository_DeleteRefreshToken(t *testing.T) {
	t.Parallel()

	must := require.New(t)
	repo, pool := newAuthRepository(t)
	ctx := context.Background()

	userID := seedUser(t, pool)
	saved := mustToken(userID, "hash-delete", time.Now().UTC().Add(time.Hour))
	must.NoError(repo.SaveRefreshToken(ctx, saved))
	must.NoError(repo.DeleteRefreshToken(ctx, saved.TokenHash))

	_, err := repo.GetRefreshTokenByHash(ctx, saved.TokenHash)
	must.ErrorIs(err, errs.ErrRefreshTokenNotFound)
}

func TestAuthRepository_DeleteExpiredRefreshTokens(t *testing.T) {
	t.Parallel()

	must := require.New(t)
	is := assert.New(t)
	repo, pool := newAuthRepository(t)
	ctx := context.Background()

	userID := seedUser(t, pool)
	now := time.Now().UTC()

	expired := mustToken(userID, "hash-expired", now.Add(-time.Hour))
	valid := mustToken(userID, "hash-valid", now.Add(24*time.Hour))

	must.NoError(repo.SaveRefreshToken(ctx, expired))
	must.NoError(repo.SaveRefreshToken(ctx, valid))

	deleted, err := repo.DeleteExpiredRefreshTokens(ctx)
	must.NoError(err)
	is.Equal(int64(1), deleted)

	_, err = repo.GetRefreshTokenByHash(ctx, expired.TokenHash)
	must.ErrorIs(err, errs.ErrRefreshTokenNotFound)

	_, err = repo.GetRefreshTokenByHash(ctx, valid.TokenHash)
	must.NoError(err)
}
