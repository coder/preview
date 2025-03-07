package preview

import (
	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	tfcontext "github.com/aquasecurity/trivy/pkg/iac/terraform/context"
	"github.com/zclconf/go-cty/cty"
)

func ParameterContextsEvalHook(input Input) func(ctx *tfcontext.Context, blocks terraform.Blocks, inputVars map[string]cty.Value) {
	return func(ctx *tfcontext.Context, blocks terraform.Blocks, inputVars map[string]cty.Value) {
		data := blocks.OfType("data")
		for _, block := range data {
			if block.TypeLabel() != "coder_parameter" {
				continue
			}

			if !block.GetAttribute("value").IsNil() {
				continue // Wow a value exists?!. This feels like a bug.
			}

			nameAttr := block.GetAttribute("name")
			nameVal := nameAttr.Value()
			if !nameVal.Type().Equals(cty.String) {
				continue // Ignore the errors at this point
			}

			name := nameVal.AsString()
			var value cty.Value
			pv, ok := input.RichParameterValue(name)
			if ok {
				// TODO: Handle non-string types
				value = cty.StringVal(pv)
			} else {
				// get the default value
				// TODO: Log any diags
				value, ok = evaluateCoderParameterDefault(block)
				if !ok {
					// no default value
					return
				}
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

func evaluateCoderParameterDefault(b *terraform.Block) (cty.Value, bool) {
	attributes := b.Attributes()

	//typeAttr, exists := attributes["type"]
	//valueType := cty.String // TODO: Default to string?
	//if exists {
	//	typeVal := typeAttr.Value()
	//	if !typeVal.Type().Equals(cty.String) || !typeVal.IsWhollyKnown() {
	//		// TODO: Mark this value somehow
	//		return cty.NilVal, nil
	//	}
	//
	//	var err error
	//	valueType, err = extract.ParameterCtyType(typeVal.AsString())
	//	if err != nil {
	//		// TODO: Mark this value somehow
	//		return cty.NilVal, nil
	//	}
	//}
	//
	////return cty.NilVal, hcl.Diagnostics{
	////	{
	////		Severity:    hcl.DiagError,
	////		Summary:     fmt.Sprintf("Decoding parameter type for %q", b.FullName()),
	////		Detail:      err.Error(),
	////		Subject:     &typeAttr.HCLAttribute().Range,
	////		Context:     &b.HCLBlock().DefRange,
	////		Expression:  typeAttr.HCLAttribute().Expr,
	////		EvalContext: b.Context().Inner(),
	////	},
	////}
	//
	//// TODO: We should support different tf types, but at present the tf
	//// schema is static. So only string is allowed
	//var val cty.Value

	def, exists := attributes["default"]
	if !exists {
		return cty.NilVal, false
	}

	v, diags := def.HCLAttribute().Expr.Value(b.Context().Inner())
	if diags.HasErrors() {
		return cty.NilVal, false
	}
	return v, true
}
