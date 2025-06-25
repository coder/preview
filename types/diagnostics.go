package types

import (
	"encoding/json"

	"github.com/hashicorp/hcl/v2"
)

const (
	// DiagnosticCodeRequired is used when a parameter value is `null`, but
	// the parameter is required.
	DiagnosticCodeRequired = "required"
)

type DiagnosticExtra struct {
	Code string `json:"code"`

	// If there was a previous extra, store it here for unwrapping.
	Wrapped any
}

var _ hcl.DiagnosticExtraUnwrapper = DiagnosticExtra{}

func (e DiagnosticExtra) UnwrapDiagnosticExtra() interface{} {
	return e.Wrapped
}

func DiagnosticCode(diag *hcl.Diagnostic, code string) *hcl.Diagnostic {
	SetDiagnosticExtra(diag, DiagnosticExtra{
		Code: code,
	})
	return diag
}

func ExtractDiagnosticExtra(diag *hcl.Diagnostic) DiagnosticExtra {
	// Zero values for a missing extra field is fine.
	extra, _ := hcl.DiagnosticExtra[DiagnosticExtra](diag)
	return extra
}

func SetDiagnosticExtra(diag *hcl.Diagnostic, extra DiagnosticExtra) {
	existing, ok := hcl.DiagnosticExtra[DiagnosticExtra](diag)
	if ok {
		// If an existing extra is present, we will keep the underlying
		// Wrapped. This is not perfect, as any parents are lost.
		// So try to avoid calling 'SetDiagnosticExtra' more than once.
		// TODO: Fix this so we maintain the parents too. Maybe use a pointer?
		extra.Wrapped = existing.Wrapped
		diag.Extra = extra
		return
	}

	// Maintain any existing extra fields.
	if diag.Extra != nil {
		extra.Wrapped = diag.Extra
	}
	diag.Extra = extra
}

// Diagnostics is a JSON friendly form of hcl.Diagnostics.
// Data is lost when doing a json marshal.
type Diagnostics hcl.Diagnostics

func (d Diagnostics) FriendlyDiagnostics() []FriendlyDiagnostic {
	cpy := make([]FriendlyDiagnostic, 0, len(d))
	for _, diag := range d {
		severity := DiagnosticSeverityError
		if diag.Severity == hcl.DiagWarning {
			severity = DiagnosticSeverityWarning
		}

		extra := ExtractDiagnosticExtra(diag)

		cpy = append(cpy, FriendlyDiagnostic{
			Severity: severity,
			Summary:  diag.Summary,
			Detail:   diag.Detail,
			Extra:    extra,
		})
	}
	return cpy
}

func (d *Diagnostics) UnmarshalJSON(data []byte) error {
	cpy := make([]FriendlyDiagnostic, 0)
	if err := json.Unmarshal(data, &cpy); err != nil {
		return err
	}

	*d = make(Diagnostics, 0, len(cpy))
	for _, diag := range cpy {
		severity := hcl.DiagError
		if diag.Severity == DiagnosticSeverityWarning {
			severity = hcl.DiagWarning
		}

		hclDiag := &hcl.Diagnostic{
			Severity: severity,
			Summary:  diag.Summary,
			Detail:   diag.Detail,
		}

		SetDiagnosticExtra(hclDiag, diag.Extra)

		*d = append(*d, hclDiag)
	}
	return nil
}

func (d Diagnostics) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.FriendlyDiagnostics())
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

	Extra DiagnosticExtra `json:"extra"`
}
