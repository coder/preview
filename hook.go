package preview

import (
	"fmt"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	tfcontext "github.com/aquasecurity/trivy/pkg/iac/terraform/context"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"

	"github.com/coder/preview/hclext"
)

func ParameterContextsEvalHook(input Input, diags hcl.Diagnostics) func(ctx *tfcontext.Context, blocks terraform.Blocks, inputVars map[string]cty.Value) {
	return func(ctx *tfcontext.Context, blocks terraform.Blocks, inputVars map[string]cty.Value) {
		data := blocks.OfType("data")
		for _, block := range data {
			if block.TypeLabel() != "coder_parameter" {
				continue
			}

			if !block.GetAttribute("value").IsNil() {
				continue // Wow a value exists?!. This feels like a bug.
			}

			name := block.NameLabel()
			var defDiags hcl.Diagnostics
			var value cty.Value
			pv, ok := input.RichParameterValue(name)
			if ok {
				// TODO: Handle non-string types
				value = pv.Value
			} else {
				// get the default value
				value, defDiags = evaluateCoderParameterDefault(block)
				diags = diags.Extend(defDiags)
			}

			// Set the default value as the 'value' attribute
			path := []string{"data"}
			path = append(path, block.Labels()...)
			path = append(path, "value")
			// The current context is in the `coder_parameter` block.
			// Use the parent context to "export" the value
			ctx.Set(value, path...)
			//block.Context().Parent().Set(value, path...)
		}
	}
}

func evaluateCoderParameterDefault(b *terraform.Block) (cty.Value, hcl.Diagnostics) {
	//if b.Label() == "" {
	//	return cty.NilVal,  errors.New("empty label - cannot resolve")
	//}

	attributes := b.Attributes()
	if attributes == nil {
		r := b.HCLBlock().Body.MissingItemRange()
		return cty.NilVal, hcl.Diagnostics{
			{
				Severity: hcl.DiagWarning,
				Summary:  "'coder_parameter' block has no attributes",
				Detail:   "No default value will be set for this paramete",
				Subject:  &r,
			},
		}
	}

	var valType cty.Type
	var defaults *typeexpr.Defaults
	// TODO: `"string"` fails, it should be `string`
	typeAttr, exists := attributes["type"]
	if exists {
		ty, def, err := hclext.DecodeVarType(typeAttr.HCLAttribute().Expr)
		if err != nil {
			return cty.NilVal, hcl.Diagnostics{
				{
					Severity:    hcl.DiagWarning,
					Summary:     fmt.Sprintf("Decoding parameter type for %q", b.FullName()),
					Detail:      err.Error(),
					Subject:     &typeAttr.HCLAttribute().Range,
					Context:     &b.HCLBlock().DefRange,
					Expression:  typeAttr.HCLAttribute().Expr,
					EvalContext: b.Context().Inner(),
				},
			}
		}
		valType = ty
		defaults = def
	} else {
		// Default to string type
		valType = cty.String
	}

	var val cty.Value

	def, exists := attributes["default"]
	if exists {
		val = def.NullableValue()
	} else {
		return cty.NilVal, nil
	}

	if valType != cty.NilType {
		if defaults != nil {
			val = defaults.Apply(val)
		}

		typedVal, err := convert.Convert(val, valType)
		if err != nil {
			return cty.NilVal, hcl.Diagnostics{
				{
					Severity:    hcl.DiagWarning,
					Summary:     "Converting default parameter value type",
					Detail:      err.Error(),
					Subject:     &def.HCLAttribute().Range,
					Context:     &b.HCLBlock().DefRange,
					Expression:  def.HCLAttribute().Expr,
					EvalContext: b.Context().Inner(),
				},
			}
		}
		return typedVal, nil
	}

	return val, nil

}
