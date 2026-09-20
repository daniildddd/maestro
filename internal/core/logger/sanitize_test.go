package logger_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/daniildddd/maestro/internal/core/logger"
)

func TestSanitize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "strips line feed",
			in:   "alice\nbob",
			want: "alicebob",
		},
		{
			name: "strips carriage return",
			in:   "alice\rbob",
			want: "alicebob",
		},
		{
			name: "strips CRLF",
			in:   "alice\r\nbob",
			want: "alicebob",
		},
		{
			name: "strips only line breaks",
			in:   "\n\r\n",
			want: "",
		},
		{
			name: "keeps clean value unchanged",
			in:   "/api/v1/users",
			want: "/api/v1/users",
		},
		{
			name: "keeps empty value empty",
			in:   "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)

			is.Equal(tt.want, logger.Sanitize(tt.in))
		})
	}
}
