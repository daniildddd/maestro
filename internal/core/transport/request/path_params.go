package request

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/errs"
)

func GetUUIDPathParam(r *http.Request, key string) (uuid.UUID, error) {
	const op = "transport.request.GetUUIDPathParam"

	param := r.PathValue(key)

	val, err := uuid.Parse(param)
	if err != nil {
		return uuid.Nil, fmt.Errorf(
			"%s: param=%q by key=%s not a valid uuid %w: %v",
			op,
			param,
			key,
			errs.ErrInvalidPathParam,
			err,
		)
	}

	return val, nil
}
