package dbcheck

import (
	"fmt"
	"strings"

	"github.com/daniildddd/maestro/internal/core/domain"
)

type configProblem struct {
	field    string
	message  string
	severity domain.CheckSeverity
}

func check(id string, severity domain.CheckSeverity, format string, args ...any) domain.ValidationCheck {
	return domain.ValidationCheck{
		ID:       id,
		Severity: severity,
		Message:  fmt.Sprintf(format, args...),
	}
}

func fieldCheck(id string, severity domain.CheckSeverity, field, format string, args ...any) domain.ValidationCheck {
	result := check(id, severity, format, args...)
	result.Field = field

	return result
}

func tableCheck(id string, severity domain.CheckSeverity, table, format string, args ...any) domain.ValidationCheck {
	result := check(id, severity, format, args...)
	result.Table = table

	return result
}

func fixCheck(source domain.ValidationCheck, fixHint string) domain.ValidationCheck {
	source.FixHint = fixHint

	return source
}

func skippedStep(id domain.StepID, reason string) domain.ValidationStep {
	return domain.ValidationStep{
		ID: id,
		Checks: []domain.ValidationCheck{{
			ID:       string(id) + ".skipped",
			Severity: domain.CheckSeveritySkipped,
			Message:  reason,
		}},
	}
}

var dbStepOrder = []domain.StepID{
	domain.StepConnection,
	domain.StepPermissions,
	domain.StepCDC,
	domain.StepTables,
}

func skippedDBSteps(after domain.StepID, reason string) []domain.ValidationStep {
	steps := make([]domain.ValidationStep, 0, len(dbStepOrder))

	skipping := false

	for _, step := range dbStepOrder {
		if step == after {
			skipping = true

			continue
		}

		if skipping {
			steps = append(steps, skippedStep(step, reason))
		}
	}

	return steps
}

func splitList(value string) []string {
	parts := strings.Split(value, ",")

	out := make([]string, 0, len(parts))

	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}

	return out
}

func isRegexFilter(entry string) bool {
	return strings.ContainsAny(entry, "*+?[](){}|^$\\")
}

const defaultSchema = "public"

func splitSchemaTable(entry string) (schema, table string) {
	if before, after, ok := strings.Cut(entry, "."); ok {
		return before, after
	}

	return defaultSchema, entry
}
