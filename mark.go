package preview

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

func markWithDiagnostic(v cty.Value, diag hcl.Diagnostics) cty.Value {
	return v.Mark(diag)
}
