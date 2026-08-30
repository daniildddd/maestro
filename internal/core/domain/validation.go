package domain

type CheckSeverity string

const (
	CheckSeverityOK      CheckSeverity = "ok"
	CheckSeverityWarning CheckSeverity = "warning"
	CheckSeverityError   CheckSeverity = "error"
	CheckSeveritySkipped CheckSeverity = "skipped"
)

type StepID string

const (
	StepConfig      StepID = "config"
	StepConnection  StepID = "connection"
	StepPermissions StepID = "permissions"
	StepCDC         StepID = "cdc"
	StepTables      StepID = "tables"
)

var AllValidationSteps = []StepID{
	StepConfig,
	StepConnection,
	StepPermissions,
	StepCDC,
	StepTables,
}

type ValidationCheck struct {
	ID       string
	Severity CheckSeverity
	Message  string
	FixHint  string
	Field    string
	Table    string
}

type ValidationStep struct {
	ID     StepID
	Status CheckSeverity
	Checks []ValidationCheck
}

type ValidationReport struct {
	Valid bool
	Steps []ValidationStep
}

func ComposeStepStatus(checks []ValidationCheck) CheckSeverity {
	hasWarning := false
	hasOK := false

	for _, check := range checks {
		if check.Severity == CheckSeverityError {
			return CheckSeverityError
		}

		if check.Severity == CheckSeverityWarning {
			hasWarning = true
		}

		if check.Severity == CheckSeverityOK {
			hasOK = true
		}
	}

	switch {
	case hasWarning:
		return CheckSeverityWarning
	case hasOK:
		return CheckSeverityOK
	default:
		return CheckSeveritySkipped
	}
}

func ComposeReport(steps []ValidationStep) ValidationReport {
	report := ValidationReport{
		Valid: true,
		Steps: steps,
	}

	for i := range report.Steps {
		report.Steps[i].Status = ComposeStepStatus(report.Steps[i].Checks)

		if report.Steps[i].Status == CheckSeverityError {
			report.Valid = false
		}
	}

	return report
}
