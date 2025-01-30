package preview

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

const (
	diagnosticsMark = "diagnostics"
)

func markWithDiagnostic(v cty.Value, diag hcl.Diagnostics) cty.Value {
	return v.Mark(diag)
}

func valueDiagnostic(v cty.Value) hcl.Diagnostics {
	x := v.Marks()
	var _ = x
	return nil
}
