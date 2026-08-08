package errs_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/daniildddd/maestro/internal/core/errs"
)

func TestAppError_Error(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	msg := "err message"
	err := &errs.AppError{
		Message: msg,
	}

	is.Equal(msg, err.Error())
}

func TestAppError_Is(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source *errs.AppError
		target error
		want   bool
	}{
		{
			name:   "same code returns true",
			source: &errs.AppError{Code: "ERR_CODE"},
			target: &errs.AppError{Code: "ERR_CODE"},
			want:   true,
		},
		{
			name:   "wrapped error matched through chain",
			source: &errs.AppError{Code: "ERR_CODE"},
			target: fmt.Errorf("wrapped: %w", &errs.AppError{Code: "ERR_CODE"}),
			want:   true,
		},
		{
			name:   "different code returns false",
			source: &errs.AppError{Code: "ERR_CODE"},
			target: &errs.AppError{Code: "ANOTHER_ERR_CODE"},
			want:   false,
		},
		{
			name:   "non-AppError target returns false",
			source: &errs.AppError{Code: "ERR_CODE"},
			target: errors.New("no app err"),
			want:   false,
		},
		{
			name:   "nil target returns false",
			source: &errs.AppError{Code: "ERR_CODE"},
			target: nil,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)

			got := tt.source.Is(tt.target)

			is.Equal(tt.want, got)
		})
	}
}
