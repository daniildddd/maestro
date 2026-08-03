package response

import (
	"encoding/json"
	"errors"
	"fmt"
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

func (rh *HTTPResponseHandler) ErrorResponse(err error) {
	rw, ok := rh.w.(*RWriter)
	if !ok {
		panic("response: underlying ResponseWriter must be *RWriter")
	}

	rw.RawErr = err

	var appErr *errs.AppError
	if errors.As(err, &appErr) {
		if appErr.HTTPStatus == http.StatusUnauthorized {
			rw.Header().Set("WWW-Authenticate", "Bearer")
		}

		rw.AppErr = appErr
		rh.JSONResponse(
			ErrorResponse{
				Code:    appErr.Code,
				Message: appErr.Message,
			},
			appErr.HTTPStatus,
		)
		return
	}

	rw.AppErr = errs.ErrInternal
	rh.errorResponse()
}

func (rh *HTTPResponseHandler) NoContent() {
	rh.w.WriteHeader(http.StatusNoContent)
}

func (rh *HTTPResponseHandler) JSONResponse(
	responseBody any,
	statusCode int,
) {
	rh.w.Header().Set("Content-Type", "application/json")

	rh.w.WriteHeader(statusCode)

	if err := json.NewEncoder(rh.w).Encode(responseBody); err != nil {
		rh.log.Error("write HTTP response", zap.Error(err))
	}
}

func (rh *HTTPResponseHandler) PanicResponse(p any) {
	rw, ok := rh.w.(*RWriter)
	if !ok {
		panic("response: underlying ResponseWriter must be *RWriter.")
	}

	err := fmt.Errorf("unexpected panic: %v", p)

	rw.RawErr = err
	rw.AppErr = errs.ErrInternal

	rh.errorResponse()
}

func (rh *HTTPResponseHandler) errorResponse() {
	errResponse := ErrorResponse{
		Code:    errs.ErrInternal.Code,
		Message: errs.ErrInternal.Message,
	}

	rh.JSONResponse(
		errResponse,
		errs.ErrInternal.HTTPStatus,
	)
}
