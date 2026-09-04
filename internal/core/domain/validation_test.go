package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComposeStepStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		checks []ValidationCheck
		want   CheckSeverity
	}{
		{
			name:   "ok",
			checks: []ValidationCheck{{Severity: CheckSeverityOK}},
			want:   CheckSeverityOK,
		},
		{
			name: "ok with warning",
			checks: []ValidationCheck{
				{Severity: CheckSeverityOK},
				{Severity: CheckSeverityWarning},
			},
			want: CheckSeverityWarning,
		},
		{
			name: "error beats warning",
			checks: []ValidationCheck{
				{Severity: CheckSeverityOK},
				{Severity: CheckSeverityWarning},
				{Severity: CheckSeverityError},
			},
			want: CheckSeverityError,
		},
		{
			name:   "only skipped",
			checks: []ValidationCheck{{Severity: CheckSeveritySkipped}},
			want:   CheckSeveritySkipped,
		},
		{
			name:   "only warning",
			checks: []ValidationCheck{{Severity: CheckSeverityWarning}},
			want:   CheckSeverityWarning,
		},
		{
			name: "empty",
			want: CheckSeveritySkipped,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, ComposeStepStatus(tt.checks))
		})
	}
}

func TestComposeReport(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		steps     []ValidationStep
		wantValid bool
	}{
		{
			name: "all checks ok means valid report",
			steps: []ValidationStep{
				{ID: StepConfig, Checks: []ValidationCheck{{Severity: CheckSeverityOK}}},
				{ID: StepConnection, Checks: []ValidationCheck{{Severity: CheckSeverityOK}}},
			},
			wantValid: true,
		},
		{
			name: "warning does not invalidate report",
			steps: []ValidationStep{
				{ID: StepConfig, Checks: []ValidationCheck{{Severity: CheckSeverityOK}}},
				{ID: StepTables, Checks: []ValidationCheck{{Severity: CheckSeverityWarning}}},
			},
			wantValid: true,
		},
		{
			name: "error invalidates report even with skipped steps",
			steps: []ValidationStep{
				{ID: StepConfig, Checks: []ValidationCheck{{Severity: CheckSeverityOK}}},
				{ID: StepCDC, Checks: []ValidationCheck{{Severity: CheckSeverityError}}},
				{ID: StepTables, Checks: []ValidationCheck{{Severity: CheckSeveritySkipped}}},
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			report := ComposeReport(tt.steps)

			assert.Equal(t, tt.wantValid, report.Valid)
			assert.Equal(t, CheckSeverityOK, report.Steps[0].Status)
		})
	}
}
