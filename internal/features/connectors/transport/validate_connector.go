package transport

import (
	"fmt"
	"net/http"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/core/transport/request"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
)

type ConnectorValidateRequest struct {
	PluginType string            `json:"plugin_type" validate:"required"`
	Config     map[string]string `json:"config"      validate:"required"`
}

type validateCheckResponse struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	FixHint  string `json:"fix_hint,omitempty"`
	Field    string `json:"field,omitempty"`
	Table    string `json:"table,omitempty"`
}

type validateStepResponse struct {
	ID     string                  `json:"id"`
	Status string                  `json:"status"`
	Checks []validateCheckResponse `json:"checks"`
}

type validateResponse struct {
	Valid bool                   `json:"valid"`
	Steps []validateStepResponse `json:"steps"`
}

func validateResponseFromDomain(report domain.ValidationReport) validateResponse {
	steps := make([]validateStepResponse, 0, len(report.Steps))

	for _, step := range report.Steps {
		checks := make([]validateCheckResponse, 0, len(step.Checks))

		for _, check := range step.Checks {
			checks = append(checks, validateCheckResponse{
				ID:       check.ID,
				Severity: string(check.Severity),
				Message:  check.Message,
				FixHint:  check.FixHint,
				Field:    check.Field,
				Table:    check.Table,
			})
		}

		steps = append(steps, validateStepResponse{
			ID:     string(step.ID),
			Status: string(step.Status),
			Checks: checks,
		})
	}

	return validateResponse{
		Valid: report.Valid,
		Steps: steps,
	}
}

func (h *ConnectorsHTTPHandler) ValidateConnector(w http.ResponseWriter, r *http.Request) {
	const op = "connectors.transport.ValidateConnector"

	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(w, log)

	var req ConnectorValidateRequest

	if err := request.DecodeAndValidate(w, r, &req); err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: decode request: %w", op, err))

		return
	}

	report, err := h.connectorsService.ValidateConnector(ctx, req.PluginType, req.Config)
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("%s: validate connector: %w", op, err))

		return
	}

	response := validateResponseFromDomain(report)

	if response.Valid {
		responseHandler.JSONResponse(response, http.StatusOK)

		return
	}

	appErr := *errs.ErrConnectorConfigInvalid
	appErr.Details = response

	responseHandler.ErrorResponse(fmt.Errorf("%s: %w", op, &appErr))
}
