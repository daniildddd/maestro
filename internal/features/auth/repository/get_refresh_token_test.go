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
	core_postgres_pool "github.com/daniildddd/maestro/internal/core/repository/postgres"
	"github.com/daniildddd/maestro/internal/features/auth/repository"
)

func TestGetRefreshTokenByHash(t *testing.T) {
	t.Parallel()

	const opTimeout = 100 * time.Millisecond

	id := uuid.New()
	userID := uuid.New()
	createdAt := time.Now().Add(-time.Hour)
	expiresAt := time.Now().Add(time.Hour)

	tests := []struct {
		name      string
		token     string
		setup     func(pool *MockPool, row *MockRow)
		wantIs    error
		wantNotIs error
		wantTok   domain.RefreshToken
	}{
		{
			name:  "success scans token by hash",
			token: "hashed-token",
			setup: func(pool *MockPool, row *MockRow) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					QueryRow(
						mock.MatchedBy(func(ctx context.Context) bool {
							deadline, ok := ctx.Deadline()
							if !ok {
								return false
							}

							return time.Until(deadline) <= opTimeout
						}),
						mock.Anything,
						[]any{"hashed-token"},
					).
					Return(row).
					Once()

				row.EXPECT().
					Scan(mock.Anything).
					Run(func(dest ...any) {
						ptrs, ok := dest[0].([]any)
						if !ok {
							return
						}

						idPtr, ok := ptrs[0].(*uuid.UUID)
						if !ok {
							return
						}

						userIDPtr, ok := ptrs[1].(*uuid.UUID)
						if !ok {
							return
						}

						hashPtr, ok := ptrs[2].(*string)
						if !ok {
							return
						}

						createdAtPtr, ok := ptrs[3].(*time.Time)
						if !ok {
							return
						}

						expiresAtPtr, ok := ptrs[4].(*time.Time)
						if !ok {
							return
						}

						*idPtr = id
						*userIDPtr = userID
						*hashPtr = "hashed-token"
						*createdAtPtr = createdAt
						*expiresAtPtr = expiresAt
					}).
					Return(nil).
					Once()
			},
			wantTok: domain.NewRefreshToken(
				id,
				userID,
				"hashed-token",
				createdAt,
				expiresAt,
			),
		},
		{
			name:  "no rows maps to ErrRefreshTokenNotFound",
			token: "hashed-token",
			setup: func(pool *MockPool, row *MockRow) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					QueryRow(mock.Anything, mock.Anything, []any{"hashed-token"}).
					Return(row).
					Once()

				row.EXPECT().
					Scan(mock.Anything).
					Return(core_postgres_pool.ErrNoRows).
					Once()
			},
			wantIs:    errs.ErrRefreshTokenNotFound,
			wantNotIs: core_postgres_pool.ErrNoRows,
		},
		{
			name:  "scan error is wrapped",
			token: "hashed-token",
			setup: func(pool *MockPool, row *MockRow) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					QueryRow(mock.Anything, mock.Anything, []any{"hashed-token"}).
					Return(row).
					Once()

				row.EXPECT().
					Scan(mock.Anything).
					Return(errs.ErrInternal).
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

			pool := NewMockPool(t)
			row := NewMockRow(t)
			tt.setup(pool, row)

			repo := repository.NewAuthRepository(pool)

			tok, err := repo.GetRefreshTokenByHash(context.Background(), tt.token)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				if tt.wantNotIs != nil {
					must.NotErrorIs(err, tt.wantNotIs)
				}

				return
			}

			must.NoError(err)
			is.Equal(tt.wantTok, tok)
		})
	}
}
