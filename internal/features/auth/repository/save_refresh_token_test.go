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

func TestSaveRefreshToken(t *testing.T) {
	t.Parallel()

	const opTimeout = 100 * time.Millisecond

	token := domain.NewRefreshToken(
		uuid.New(),
		uuid.New(),
		"hashed-token",
		time.Now().Add(-time.Hour),
		time.Now().Add(time.Hour),
	)

	tests := []struct {
		name   string
		token  domain.RefreshToken
		setup  func(pool *MockPool)
		wantIs error
	}{
		{
			name:  "success saves token within op timeout",
			token: token,
			setup: func(pool *MockPool) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Exec(
						mock.MatchedBy(func(ctx context.Context) bool {
							deadline, ok := ctx.Deadline()
							if !ok {
								return false
							}

							return time.Until(deadline) <= opTimeout
						}),
						mock.Anything,
						[]any{
							token.Id,
							token.UserId,
							token.TokenHash,
							token.CreatedAt,
							token.ExpiresAt,
						},
					).
					Return(nil, nil).
					Once()
			},
		},
		{
			name:  "foreign key violation is wrapped with FK violation",
			token: token,
			setup: func(pool *MockPool) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Exec(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, core_postgres_pool.ErrViolatesForeignKey).
					Once()
			},
			wantIs: core_postgres_pool.ErrViolatesForeignKey,
		},
		{
			name:  "exec error is wrapped",
			token: token,
			setup: func(pool *MockPool) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Exec(mock.Anything, mock.Anything, mock.Anything).
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

			pool := NewMockPool(t)
			tt.setup(pool)

			repo := repository.NewAuthRepository(pool)

			err := repo.SaveRefreshToken(context.Background(), tt.token)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
		})
	}
}
