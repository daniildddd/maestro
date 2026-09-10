package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/features/auth/repository"
)

func TestAuthRepository_DeleteExpiredRefreshTokens(t *testing.T) {
	t.Parallel()

	const opTimeout = 100 * time.Millisecond

	tests := []struct {
		name      string
		setup     func(pool *MockPool)
		wantIs    error
		wantCount int64
	}{
		{
			name: "success returns deleted count",
			setup: func(pool *MockPool) {
				pool.EXPECT().
					OpTimeout().
					Return(opTimeout).
					Once()

				pool.EXPECT().
					Exec(mock.Anything, mock.Anything, mock.Anything).
					Return(stubCommandTag{affected: 5}, nil).
					Once()
			},
			wantCount: 5,
		},
		{
			name: "exec error is wrapped",
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

			count, err := repo.DeleteExpiredRefreshTokens(context.Background())

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			is.Equal(tt.wantCount, count)
		})
	}
}
