package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/errs"
)

func (s *ConnectorsService) ValidateConnector(
	ctx context.Context,
	pluginType string,
	config map[string]string,
) (domain.ValidationReport, error) {
	const op = "connectors.service.ValidateConnector"

	checker := s.checkerFor(pluginType)
	if checker == nil {
		appErr := *errs.ErrConnectorValidationUnsupported
		appErr.Message = fmt.Sprintf("validation is not supported for connector type %q", pluginType)

		return domain.ValidationReport{}, fmt.Errorf("%s: %w", op, &appErr)
	}

	stepReports := make([]domain.ValidationStep, 0, len(domain.AllValidationSteps))

	configStep, err := s.configStep(ctx, pluginType, config)
	if err != nil {
		return domain.ValidationReport{}, fmt.Errorf("%s: %w", op, err)
	}

	stepReports = append(stepReports, configStep)

	dbSteps, checkErr := checker.Check(ctx, config)
	if checkErr != nil {
		return domain.ValidationReport{}, fmt.Errorf("%s: %w", op, checkErr)
	}

	stepReports = append(stepReports, dbSteps...)

	return domain.ComposeReport(stepReports), nil
}

func (s *ConnectorsService) checkerFor(class string) dbChecker {
	for _, checker := range s.checkers {
		if checker.Match(class) {
			return checker
		}
	}

	return nil
}

func (s *ConnectorsService) configStep(
	ctx context.Context,
	pluginType string,
	config map[string]string,
) (domain.ValidationStep, error) {
	const op = "connectors.service.configStep"

	checks, err := s.connectors.ValidateConfig(ctx, pluginType, config)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrConnectorPluginNotFound):
			return domain.ValidationStep{}, errs.ErrConnectorPluginNotFound
		case errors.Is(err, domain.ErrKafkaConnectUnavailable):
			return domain.ValidationStep{}, errs.ErrKafkaConnectUnavailable
		default:
			return domain.ValidationStep{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	return domain.ValidationStep{ID: domain.StepConfig, Checks: checks}, nil
}
