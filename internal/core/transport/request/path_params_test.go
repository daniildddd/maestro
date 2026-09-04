package request_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/transport/request"
)

func TestGetIntPathParam(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     string
		value   string
		wantVal int
		wantIs  error
	}{
		{
			name:    "valid integer returns value",
			key:     "id",
			value:   "42",
			wantVal: 42,
		},
		{
			name:    "negative integer returns value",
			key:     "id",
			value:   "-7",
			wantVal: -7,
		},
		{
			name:   "non-integer returns ErrInvalidPathParam",
			key:    "id",
			value:  "abc",
			wantIs: errs.ErrInvalidPathParam,
		},
		{
			name:   "float value returns ErrInvalidPathParam",
			key:    "id",
			value:  "3.14",
			wantIs: errs.ErrInvalidPathParam,
		},
		{
			name:   "empty param value returns ErrInvalidPathParam",
			key:    "id",
			value:  "",
			wantIs: errs.ErrInvalidPathParam,
		},
		{
			name:   "unknown key returns ErrInvalidPathParam",
			key:    "other",
			value:  "42",
			wantIs: errs.ErrInvalidPathParam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			must := require.New(t)

			req := httptest.NewRequest(http.MethodGet, "/tasks/{id}", http.NoBody)
			req.SetPathValue("id", tt.value)

			got, err := request.GetIntPathParam(req, tt.key)

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

func TestGetStringPathParam(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     string
		value   string
		wantVal string
		wantIs  error
	}{
		{
			name:    "valid string returns value",
			key:     "name",
			value:   "pg-connector",
			wantVal: "pg-connector",
		},
		{
			name:   "empty param value returns ErrInvalidPathParam",
			key:    "name",
			value:  "",
			wantIs: errs.ErrInvalidPathParam,
		},
		{
			name:   "unknown key returns ErrInvalidPathParam",
			key:    "other",
			value:  "pg-connector",
			wantIs: errs.ErrInvalidPathParam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			must := require.New(t)

			req := httptest.NewRequest(http.MethodGet, "/connectors/{name}", http.NoBody)
			req.SetPathValue("name", tt.value)

			got, err := request.GetStringPathParam(req, tt.key)

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

func TestGetUUIDPathParam(t *testing.T) {
	t.Parallel()

	want := uuid.New()

	tests := []struct {
		name    string
		key     string
		value   string
		wantVal uuid.UUID
		wantIs  error
	}{
		{
			name:    "valid uuid returns value",
			key:     "id",
			value:   want.String(),
			wantVal: want,
		},
		{
			name:   "non-uuid returns ErrInvalidPathParam",
			key:    "id",
			value:  "not-a-uuid",
			wantIs: errs.ErrInvalidPathParam,
		},
		{
			name:   "nil uuid returns ErrInvalidPathParam",
			key:    "id",
			value:  "00000000-0000-0000-0000-000000000000",
			wantIs: errs.ErrInvalidPathParam,
		},
		{
			name:   "empty param value returns ErrInvalidPathParam",
			key:    "id",
			value:  "",
			wantIs: errs.ErrInvalidPathParam,
		},
		{
			name:   "unknown key returns ErrInvalidPathParam",
			key:    "other",
			value:  want.String(),
			wantIs: errs.ErrInvalidPathParam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			req := httptest.NewRequest(http.MethodGet, "/users/{id}", http.NoBody)
			req.SetPathValue("id", tt.value)

			got, err := request.GetUUIDPathParam(req, tt.key)

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
