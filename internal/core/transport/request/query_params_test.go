package request_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/transport/request"
)

func TestGetIntQueryParam(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		query   string
		key     string
		wantVal int
		wantIs  error
	}{
		{
			name:    "valid integer returns value",
			query:   "page=42",
			key:     "page",
			wantVal: 42,
		},
		{
			name:    "negative integer returns value",
			query:   "offset=-5",
			key:     "offset",
			wantVal: -5,
		},
		{
			name:  "unknown key returns zero without error",
			query: "page=42",
			key:   "limit",
		},
		{
			name:    "empty param value returns zero without error",
			query:   "page=",
			key:     "page",
			wantVal: 0,
		},
		{
			name:   "non-integer returns ErrInvalidCredentials",
			query:  "page=abc",
			key:    "page",
			wantIs: errs.ErrInvalidCredentials,
		},
		{
			name:   "overflowing integer returns ErrInvalidCredentials",
			query:  "page=999999999999999999999",
			key:    "page",
			wantIs: errs.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			req := httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)

			got, err := request.GetIntQueryParam(req, tt.key)

			if tt.wantIs != nil {
				must.Error(err)
				is.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			is.Equal(tt.wantVal, got)
		})
	}
}
