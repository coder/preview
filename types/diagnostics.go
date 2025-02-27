package types

import (
	"encoding/json"

	"github.com/hashicorp/hcl/v2"
)

// Diagnostics is a JSON friendly form of hcl.Diagnostics.
// Data is lost when doing a json marshal.
type Diagnostics hcl.Diagnostics

func (d Diagnostics) MarshalJSON() ([]byte, error) {
	cpy := make([]FriendlyDiagnostic, 0, len(d))
	for _, diag := range d {
		severity := DiagnosticSeverityError
		if diag.Severity == hcl.DiagWarning {
			severity = DiagnosticSeverityWarning
		}

		cpy = append(cpy, FriendlyDiagnostic{
			Severity: severity,
			Summary:  diag.Summary,
			Detail:   diag.Detail,
		})
	}
	return json.Marshal(cpy)
}

type DiagnosticSeverityString string

const (
	DiagnosticSeverityError   DiagnosticSeverityString = "error"
	DiagnosticSeverityWarning DiagnosticSeverityString = "warning"
)

type FriendlyDiagnostic struct {
	Severity DiagnosticSeverityString `json:"severity"`
	Summary  string                   `json:"summary"`
	Detail   string                   `json:"detail"`
}
