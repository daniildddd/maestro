package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/daniildddd/maestro/internal/core/errs"
)

const maxBodyBytes = 1 << 20 // 1 MB

var requestValidator = validator.New()

type validatable interface {
	Validate() error
}

func DecodeAndValidate(
	w http.ResponseWriter,
	r *http.Request,
	dest any,
) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return fmt.Errorf(
			"parse content type: %w: %w",
			errs.ErrInvalidRequestBody,
			err,
		)
	}

	if mediaType != "application/json" {
		return fmt.Errorf(
			"invalid content type: %w: %v",
			errs.ErrInvalidRequestBody,
			mediaType,
		)
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err = dec.Decode(dest); err != nil {
		var maxBytesErr *http.MaxBytesError

		if errors.As(err, &maxBytesErr) {
			return fmt.Errorf(
				"request body too large: %w: %w",
				errs.ErrInvalidRequestBody,
				err,
			)
		}

		return fmt.Errorf(
			"decode json: %w: %w",
			errs.ErrInvalidRequestBody,
			err,
		)
	}

	v, ok := dest.(validatable)
	if ok {
		err = v.Validate()
	} else {
		err = requestValidator.Struct(dest)
	}

	if err != nil {
		return fmt.Errorf(
			"request validation: %w: %w",
			errs.ErrValidationFailed,
			err,
		)
	}

	return nil
}
