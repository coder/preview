package preview

import (
	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/coder/preview/extract"
	"github.com/coder/preview/hclext"
	"github.com/coder/preview/types"
)

func RichParameters(modules terraform.Modules) ([]types.Parameter, hcl.Diagnostics) {
	diags := make(hcl.Diagnostics, 0)
	params := make([]types.Parameter, 0)

	for _, mod := range modules {
		blocks := mod.GetDatasByType(types.BlockTypeParameter)
		for _, block := range blocks {
			param, pDiags := extract.ParameterFromBlock(block)
			if len(pDiags) > 0 {
				diags = diags.Extend(pDiags)
			}

			if !pDiags.HasErrors() {
				params = append(params, param)
			}
		}
	}

	return params, diags
}

func paramOption(block *terraform.Block) (*types.ParameterOption, hcl.Diagnostics) {
	p := newAttributeParser(block)
	opt := &types.ParameterOption{
		Name:        p.attr("name").required().string(),
		Description: p.attr("description").string(),
		// Does it need to be a string?
		Value: p.attr("value").required().tryString(),
		Icon:  p.attr("icon").string(),
	}
	if p.diags.HasErrors() {
		return nil, p.diags
	}
	return opt, nil
}

func richParameterValue(block *terraform.Block) cty.Value {
	// Find the value of the parameter from the context.
	paramPath := append([]string{"data"}, block.Labels()...)
	valueRef := hclext.ScopeTraversalExpr(append(paramPath, "value")...)
	paramValue, diags := valueRef.Value(block.Context().Inner())
	if diags != nil && diags.HasErrors() {
		for _, diag := range diags {
			b := block.HCLBlock().Body.MissingItemRange()
			diag.Subject = &b
		}

		// TODO: Figure out what to do with these diagnostics
		var _ = diags
		return cty.UnknownVal(cty.NilType)
	}

	return paramValue
}
