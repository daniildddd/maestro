package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func TestNewTokenPair(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	expiresAt := time.Now().Add(time.Hour)

	got := domain.NewTokenPair(
		"access-abc-123",
		"refresh-xyz-789",
		expiresAt,
	)

	is.Equal("access-abc-123", got.AccessToken)
	is.Equal("refresh-xyz-789", got.RefreshToken)
	is.True(got.ExpiresAt.Equal(expiresAt))
}
