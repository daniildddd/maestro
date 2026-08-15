package request_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/transport/request"
)

type userReq struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type customValidatableReq struct {
	Value string `json:"value"`
}

func (r *customValidatableReq) Validate() error {
	if r.Value != "ok" {
		return errors.New("value must be ok")
	}

	return nil
}

func TestDecodeAndValidate(t *testing.T) {
	t.Parallel()

	t.Run("valid JSON decodes into struct with tags", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		dest := &userReq{}
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"username":"alice","password":"secret123"}`))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		err := request.DecodeAndValidate(rec, req, dest)

		must.NoError(err)
		must.Equal("alice", dest.Username)
		must.Equal("secret123", dest.Password)
	})

	t.Run("valid JSON decodes into custom validatable", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		dest := &customValidatableReq{}
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"value":"ok"}`))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		err := request.DecodeAndValidate(rec, req, dest)

		must.NoError(err)
		must.Equal("ok", dest.Value)
	})

	const testMaxBodyBytes = 1 << 20 // must match maxBodyBytes in decode.go

	tests := []struct {
		name        string
		contentType string
		body        string
		dest        any
		wantErr     error
	}{
		{
			name:        "missing Content-Type returns ErrInvalidContentType",
			contentType: "",
			body:        `{"username":"alice","password":"secret123"}`,
			dest:        &userReq{},
			wantErr:     errs.ErrInvalidContentType,
		},
		{
			name:        "invalid Content-Type returns ErrInvalidContentType",
			contentType: "text/plain",
			body:        `{"username":"alice","password":"secret123"}`,
			dest:        &userReq{},
			wantErr:     errs.ErrInvalidContentType,
		},
		{
			name:        "invalid JSON syntax returns ErrInvalidRequestBody",
			contentType: "application/json",
			body:        `{invalid`,
			dest:        &userReq{},
			wantErr:     errs.ErrInvalidRequestBody,
		},
		{
			name:        "unknown field returns ErrInvalidRequestBody",
			contentType: "application/json",
			body:        `{"username":"alice","password":"secret123","extra":1}`,
			dest:        &userReq{},
			wantErr:     errs.ErrInvalidRequestBody,
		},
		{
			name:        "body exceeds limit returns ErrInvalidRequestBody",
			contentType: "application/json",
			body:        `{"data":"` + strings.Repeat("a", testMaxBodyBytes+10) + `"}`,
			dest: &struct {
				Data string `json:"data"`
			}{},
			wantErr: errs.ErrInvalidRequestBody,
		},
		{
			name:        "required field empty returns ErrValidationFailed",
			contentType: "application/json",
			body:        `{"username":"","password":""}`,
			dest:        &userReq{},
			wantErr:     errs.ErrValidationFailed,
		},
		{
			name:        "short field returns ErrValidationFailed",
			contentType: "application/json",
			body:        `{"username":"ab","password":"short"}`,
			dest:        &userReq{},
			wantErr:     errs.ErrValidationFailed,
		},
		{
			name:        "custom Validate fails returns ErrValidationFailed",
			contentType: "application/json",
			body:        `{"value":"bad"}`,
			dest:        &customValidatableReq{},
			wantErr:     errs.ErrValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			rec := httptest.NewRecorder()

			err := request.DecodeAndValidate(rec, req, tt.dest)

			must.Error(err)
			must.ErrorIs(err, tt.wantErr)
		})
	}
}
