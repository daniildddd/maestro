package response

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/logger"
	"go.uber.org/zap"
)

type HTTPResponseHandler struct {
	w   http.ResponseWriter
	log *logger.Logger
}

func NewHTTPResponseHandler(
	w http.ResponseWriter,
	log *logger.Logger,
) *HTTPResponseHandler {
	return &HTTPResponseHandler{
		w:   w,
		log: log,
	}
}

func (rh *HTTPResponseHandler) ErrorResponse(
	err error,
	msg string,
) {
	var (
		statusCode int
		codeError  string
		logFunc    func(string, ...zap.Field)
	)

	switch {
	case errors.Is(err, errs.ErrExpiredRefreshToken):
		statusCode = http.StatusUnauthorized
		logFunc = rh.log.Warn
		codeError = codeInvalidRefreshToken

	case errors.Is(err, errs.ErrInvalidRefreshToken):
		statusCode = http.StatusUnauthorized
		logFunc = rh.log.Warn
		codeError = codeInvalidRefreshToken

	case errors.Is(err, errs.ErrMissingRefreshToken):
		statusCode = http.StatusUnauthorized
		logFunc = rh.log.Warn
		codeError = codeInvalidRefreshToken

	case errors.Is(err, errs.ErrInvalidRequestBody):
		statusCode = http.StatusBadRequest
		logFunc = rh.log.Warn
		codeError = codeInvalidRequestBody

	case errors.Is(err, errs.ErrInvalidCredentials):
		statusCode = http.StatusUnauthorized
		logFunc = rh.log.Info
		codeError = codeInvalidCredentials

	case errors.Is(err, errs.ErrValidationFailed):
		statusCode = http.StatusBadRequest
		logFunc = rh.log.Info
		codeError = codeInvalidRequestBody

	default:
		statusCode = http.StatusInternalServerError
		logFunc = rh.log.Error
		codeError = codeInternal
	}

	logFunc(msg, zap.Error(err))

	rh.w.WriteHeader(statusCode)

	rh.w.Header().Set("Content-Type", "application/json")

	errorResponse := ErrorResponse{
		Code:    codeError,
		Message: msg,
	}

	if err = json.NewEncoder(rh.w).Encode(errorResponse); err != nil {
		rh.log.Error("write HTTP response")
	}
}

func (rh *HTTPResponseHandler) JSONResponse(
	responseBody any,
	statusCode int,
) {
	rh.w.WriteHeader(statusCode)

	if err := json.NewEncoder(rh.w).Encode(responseBody); err != nil {
		rh.log.Error("write HTTP response", zap.Error(err))
	}
}
