package web

import "github.com/hashicorp/hcl/v2"

type DiagnosticSeverity string

const (
	DiagnosticSeverityError   DiagnosticSeverity = "error"
	DiagnosticSeverityWarning DiagnosticSeverity = "warning"
)

type Diagnostics []Diagnostic

type Diagnostic struct {
	Severity DiagnosticSeverity `json:"severity"`
	Summary  string             `json:"summary"`
	Detail   string             `json:"detail"`
}

func FromHCLDiagnostics(diagnostics hcl.Diagnostics) Diagnostics {
	var reqDiagnostics Diagnostics
	for _, d := range diagnostics {
		reqDiagnostics = append(reqDiagnostics, FromHCLDiagnostic(d))
	}
	return reqDiagnostics
}

func FromHCLDiagnostic(d *hcl.Diagnostic) Diagnostic {
	sev := DiagnosticSeverityError
	if d.Severity == hcl.DiagWarning {
		sev = DiagnosticSeverityWarning
	}
	return Diagnostic{
		Severity: sev,
		Summary:  d.Summary,
		Detail:   d.Detail,
	}
}
