package hclext

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func DecodeVarType(exp hcl.Expression) (cty.Type, *typeexpr.Defaults, error) {
	// This block converts the string literals "string" -> string
	// Coder used to allow literal strings, instead of types as keywords. So
	// we have to handle these cases for backwards compatibility.
	if tpl, ok := exp.(*hclsyntax.TemplateExpr); ok && len(tpl.Parts) == 1 {
		if lit, ok := tpl.Parts[0].(*hclsyntax.LiteralValueExpr); ok && lit.Val.Type() == cty.String {
			keyword := lit.Val.AsString()

			exp = &hclsyntax.ScopeTraversalExpr{
				Traversal: []hcl.Traverser{
					hcl.TraverseRoot{
						Name:     keyword,
						SrcRange: exp.Range(),
					},
				},
				SrcRange: exp.Range(),
			}
		}
	}

	// Special-case the shortcuts for list(any) and map(any) which aren't hcl.
	switch hcl.ExprAsKeyword(exp) {
	case "list":
		return cty.List(cty.DynamicPseudoType), nil, nil
	case "map":
		return cty.Map(cty.DynamicPseudoType), nil, nil
	}

	t, def, diag := typeexpr.TypeConstraintWithDefaults(exp)
	if diag.HasErrors() {
		return cty.NilType, nil, diag
	}
	return t, def, nil
}
