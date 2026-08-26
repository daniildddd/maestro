package request

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/daniildddd/maestro/internal/core/errs"
)

func GetIntPathParam(r *http.Request, key string) (int, error) {
	const op = "transport.request.GetIntPathParam"

	param := r.PathValue(key)

	val, err := strconv.Atoi(param)
	if err != nil {
		return 0, fmt.Errorf(
			"%s: param=%q by key=%s not a valid integer: %w: %v",
			op,
			param,
			key,
			errs.ErrInvalidPathParam,
			err,
		)
	}

	return val, nil
}

func GetStringPathParam(r *http.Request, key string) (string, error) {
	const op = "transport.request.GetStringPathParam"

	param := r.PathValue(key)
	if param == "" {
		return "", fmt.Errorf("%s: param by key=%s is empty: %w", op, key, errs.ErrInvalidPathParam)
	}

	return param, nil
}

func GetUUIDPathParam(r *http.Request, key string) (uuid.UUID, error) {
	const op = "transport.request.GetUUIDPathParam"

	param := r.PathValue(key)

	val, err := uuid.Parse(param)
	if err != nil {
		return uuid.Nil, fmt.Errorf(
			"%s: param=%q by key=%s not a valid uuid: %w: %v",
			op,
			param,
			key,
			errs.ErrInvalidPathParam,
			err,
		)
	}

	if val == uuid.Nil {
		return uuid.Nil, fmt.Errorf(
			"%s: param=%q by key=%s is nil uuid: %w",
			op,
			param,
			key,
			errs.ErrInvalidPathParam,
		)
	}

	return val, nil
}
