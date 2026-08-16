package request

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/daniildddd/maestro/internal/core/errs"
)

func GetIntQueryParam(r *http.Request, key string) (int, error) {
	const op = "transport.request.GetIntQueryParam"

	param := r.URL.Query().Get(key)
	if param == "" {
		return 0, nil
	}

	val, err := strconv.Atoi(param)
	if err != nil {
		return 0, fmt.Errorf(
			"%s: param=%s by key=%s not a valid integer %w: %v",
			op,
			param,
			key,
			errs.ErrInvalidCredentials,
			err,
		)
	}

	return val, nil
}
