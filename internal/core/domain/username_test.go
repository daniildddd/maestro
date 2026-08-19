package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func TestValidateUsername(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		username string
		wantErr  error
	}{
		{
			name:     "valid username",
			username: "alice",
			wantErr:  nil,
		},
		{
			name:     "valid username with digits and separators",
			username: "user_name-1",
			wantErr:  nil,
		},
		{
			name:     "minimum length accepted",
			username: "abc",
			wantErr:  nil,
		},
		{
			name:     "maximum length accepted",
			username: strings.Repeat("a", 32),
			wantErr:  nil,
		},
		{
			name:     "empty username rejected",
			username: "",
			wantErr:  domain.ErrInvalidUsername,
		},
		{
			name:     "username too short rejected",
			username: "ab",
			wantErr:  domain.ErrInvalidUsername,
		},
		{
			name:     "username too long rejected",
			username: strings.Repeat("a", 33),
			wantErr:  domain.ErrInvalidUsername,
		},
		{
			name:     "username with space rejected",
			username: "bad name",
			wantErr:  domain.ErrInvalidUsername,
		},
		{
			name:     "username with dot rejected",
			username: "user.name",
			wantErr:  domain.ErrInvalidUsername,
		},
		{
			name:     "username with non-ascii rejected",
			username: "пользователь",
			wantErr:  domain.ErrInvalidUsername,
		},
		{
			name:     "username with at sign rejected",
			username: "user@mail",
			wantErr:  domain.ErrInvalidUsername,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			err := domain.ValidateUsername(tt.username)

			if tt.wantErr != nil {
				must.ErrorIs(err, tt.wantErr)

				return
			}

			must.NoError(err)
		})
	}
}
