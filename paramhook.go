package preview

import (
	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	tfcontext "github.com/aquasecurity/trivy/pkg/iac/terraform/context"
	"github.com/zclconf/go-cty/cty"

	"github.com/coder/preview/hclext"
)

// parameterContextsEvalHook is called in a loop, so if parameters affect
// other parameters, this can solve the problem 1 "evaluation" at a time.
//
// Omitting to set a default value is OK, as long as at least 1 parameter
// is resolvable. The resolvable parameter will be accessible on the next
// iteration.
func parameterContextsEvalHook(input Input) func(ctx *tfcontext.Context, blocks terraform.Blocks, inputVars map[string]cty.Value) {
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
					// the default value cannot be resolved, so do not
					// set anything.
					continue
				}
			}

			path := []string{
				"data",
				"coder_parameter",
				block.Reference().NameLabel(),
			}
			existing := ctx.Get(path...)
			obj, ok := mergeParamInstanceValues(block, existing, value)
			if !ok {
				continue
			}
			ctx.Set(obj, path...)
		}
	}
}

func mergeParamInstanceValues(b *terraform.Block, existing cty.Value, value cty.Value) (cty.Value, bool) {
	if existing.IsNull() {
		return existing, false
	}

	ref := b.Reference()
	key := ref.RawKey()

	switch {
	case key.Type().Equals(cty.Number) && b.GetAttribute("count") != nil:
		if !existing.Type().IsTupleType() {
			return existing, false
		}

		idx, _ := key.AsBigFloat().Int64()
		elem := existing.Index(key)
		if elem.IsNull() || !elem.IsKnown() {
			return existing, false
		}

		obj, ok := setObjectField(elem, "value", value)
		if !ok {
			return existing, false
		}

		return hclext.InsertTupleElement(existing, int(idx), obj), true
	case isForEachKey(key) && b.GetAttribute("for_each") != nil:
		keyStr := ref.Key()
		if !existing.Type().IsObjectType() {
			return existing, false
		}

		if !existing.CanIterateElements() {
			return existing, false
		}

		instances := existing.AsValueMap()
		if instances == nil {
			return existing, false
		}

		instance, ok := instances[keyStr]
		if !ok {
			return existing, false
		}

		instance, ok = setObjectField(instance, "value", value)
		if !ok {
			return existing, false
		}

		instances[keyStr] = instance
		return cty.ObjectVal(instances), true

	default:
		obj, ok := setObjectField(existing, "value", value)
		if !ok {
			return existing, false
		}
		return obj, true
	}
}

func setObjectField(object cty.Value, field string, value cty.Value) (cty.Value, bool) {
	if object.IsNull() {
		return object, false
	}

	if !object.Type().IsObjectType() {
		return object, false
	}

	if !object.CanIterateElements() {
		return object, false
	}

	instances := object.AsValueMap()
	if instances == nil {
		return object, false
	}

	instances[field] = value
	return cty.ObjectVal(instances), true
}

func isForEachKey(key cty.Value) bool {
	return key.Type().Equals(cty.Number) || key.Type().Equals(cty.String)
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
