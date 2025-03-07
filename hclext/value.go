package hclext

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

func Value(expr hcl.Expression, ctx *hcl.EvalContext) cty.Value {
	val, diags := expr.Value(ctx)
	if diags.HasErrors() {
		// Should we do something?
	}
	return val
}
