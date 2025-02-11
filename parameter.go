package preview

import (
	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/coder/preview/hclext"
	"github.com/coder/preview/types"
)

func RichParameters(modules terraform.Modules) ([]types.Parameter, hcl.Diagnostics) {
	diags := make(hcl.Diagnostics, 0)
	params := make([]types.Parameter, 0)

	for _, mod := range modules {
		blocks := mod.GetDatasByType("coder_parameter")
		for _, block := range blocks {
			p := newAttributeParser(block)

			var paramOptions []*types.RichParameterOption
			optionBlocks := block.GetBlocks("option")
			for _, optionBlock := range optionBlocks {
				option, optDiags := paramOption(optionBlock)
				if optDiags.HasErrors() {
					// Add the error and continue
					diags = diags.Extend(optDiags)
					continue
				}
				paramOptions = append(paramOptions, option)
			}

			// Find the value of the parameter from the context.
			paramValue := richParameterValue(block)
			var defVal string

			// default can be nil if the references are not resolved.
			def := block.GetAttribute("default").Value()
			if def.Equals(cty.NilVal).True() {
				defVal = "<nil>"
			} else if !def.IsKnown() {
				defVal = "<unknown>"
			} else {
				defVal = p.attr("default").string()
			}

			param := types.Parameter{
				Value: types.ParameterValue{
					Value: paramValue,
				},
				RichParameter: types.RichParameter{
					BlockName:    block.Labels()[1],
					Name:         p.attr("name").required().string(),
					Description:  p.attr("description").string(),
					Type:         "",
					Mutable:      false,
					DefaultValue: defVal,
					Icon:         p.attr("icon").string(),
					Options:      paramOptions,
					Validation:   nil,
					Required:     false,
					DisplayName:  "",
					Order:        0,
					Ephemeral:    false,
				},
			}
			diags = diags.Extend(p.diags)
			if p.diags.HasErrors() {
				continue
			}
			params = append(params, param)
		}
	}

	return params, diags
}

func paramOption(block *terraform.Block) (*types.RichParameterOption, hcl.Diagnostics) {
	p := newAttributeParser(block)
	opt := &types.RichParameterOption{
		Name:        p.attr("name").required().string(),
		Description: p.attr("description").string(),
		// Does it need to be a string?
		Value: p.attr("value").required().string(),
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
