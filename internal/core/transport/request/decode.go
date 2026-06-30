package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/errs"
)

const maxBodyBytes = 1 << 20 // 1 MB

type Validator interface {
	Struct(s any) error
}

func DecodeAndValidate(
	w http.ResponseWriter,
	r *http.Request,
	validator Validator,
	dest any,
) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return fmt.Errorf(
			"parse content type: %w: %w",
			err,
			errs.ErrInvalidRequestBody,
		)
	}

	if mediaType != "application/json" {
		return fmt.Errorf(
			"invalid content type: %q: %w",
			mediaType,
			errs.ErrInvalidRequestBody,
		)
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dest); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return fmt.Errorf(
				"request body too large: %w: %w",
				err,
				errs.ErrInvalidRequestBody,
			)
		}

		return fmt.Errorf(
			"decode json: %w: %w",
			err,
			errs.ErrInvalidRequestBody,
		)
	}

	if err := validator.Struct(dest); err != nil {
		return fmt.Errorf(
			"request validation: %w: %w",
			err,
			errs.ErrValidationFailed,
		)
	}

	return nil
}
