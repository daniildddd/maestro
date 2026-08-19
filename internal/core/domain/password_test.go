package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func TestValidatePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "valid password",
			password: "secret123",
			wantErr:  nil,
		},
		{
			name:     "valid password with spaces",
			password: "my pass phrase",
			wantErr:  nil,
		},
		//nolint:gosec // test fixture: unicode password string, not a real credential
		{
			name:     "valid password with unicode",
			password: "секретный пароль",
			wantErr:  nil,
		},
		{
			name:     "valid password exactly 8 chars",
			password: "12345678",
			wantErr:  nil,
		},
		{
			name:     "valid password exactly 128 chars",
			password: strings.Repeat("a", 128),
			wantErr:  nil,
		},
		{
			name:     "password too short rejected",
			password: "1234567",
			wantErr:  domain.ErrInvalidPassword,
		},
		{
			name:     "password too long rejected",
			password: strings.Repeat("a", 129),
			wantErr:  domain.ErrInvalidPassword,
		},
		{
			name:     "empty password rejected",
			password: "",
			wantErr:  domain.ErrInvalidPassword,
		},
		{
			name:     "password with newline rejected",
			password: "secret\n123",
			wantErr:  domain.ErrInvalidPassword,
		},
		{
			name:     "password with tab rejected",
			password: "secret\t123",
			wantErr:  domain.ErrInvalidPassword,
		},
		{
			name:     "password with del char rejected",
			password: "secret\x7F123",
			wantErr:  domain.ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			err := domain.ValidatePassword(tt.password)

			if tt.wantErr != nil {
				must.ErrorIs(err, tt.wantErr)

				return
			}

			must.NoError(err)
		})
	}
}
