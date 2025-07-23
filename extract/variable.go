package extract

import (
	"fmt"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"

	"github.com/coder/preview/types"
)

// VariableFromBlock extracts a terraform variable, but not its final resolved value.
// code taken mostly from https://github.com/aquasecurity/trivy/blob/main/pkg/iac/scanners/terraform/parser/evaluator.go#L479
func VariableFromBlock(block *terraform.Block) (tfVar types.Variable) {
	defer func() {
		// Extra safety mechanism to ensure that if a panic occurs, we do not break
		// everything else.
		if r := recover(); r != nil {
			tfVar = types.Variable{
				Name: block.Label(),
				Diagnostics: types.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  "Panic occurred in extracting variable. This should not happen, please report this to Coder.",
						Detail:   fmt.Sprintf("panic in variable extract: %+v", r),
					},
				},
			}
		}
	}()

	attributes := block.Attributes()

	var valType cty.Type
	var defaults *typeexpr.Defaults

	if typeAttr, exists := attributes["type"]; exists {
		ty, def, err := typeAttr.DecodeVarType()
		if err != nil {
			var subject hcl.Range
			if typeAttr.HCLAttribute() != nil {
				subject = typeAttr.HCLAttribute().Range
			}
			return types.Variable{
				Name: block.Label(),
				Diagnostics: types.Diagnostics{&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Failed to decode variable type for " + block.Label(),
					Detail:   err.Error(),
					Subject:  &subject,
				}},
			}
		}
		valType = ty
		defaults = def
	}

	var val cty.Value
	var defSubject hcl.Range
	if def, exists := attributes["default"]; exists {
		val = def.NullableValue()
		defSubject = def.HCLAttribute().Range
	}

	if valType != cty.NilType {
		// TODO: If this code ever extracts the actual value of the variable,
		// then we need to source the value from that, rather than the default.
		if defaults != nil {
			val = defaults.Apply(val)
		}

		canConvert := !val.IsNull() && val.IsWhollyKnown() && valType != cty.NilType

		if canConvert {
			typedVal, err := convert.Convert(val, valType)
			if err != nil {
				return types.Variable{
					Name: block.Label(),
					Diagnostics: types.Diagnostics{&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary: fmt.Sprintf("Failed to convert variable default value to type %q for %q",
							valType.FriendlyNameForConstraint(), block.Label()),
						Detail:  err.Error(),
						Subject: &defSubject,
					}},
				}
			}
			val = typedVal
		}
	} else {
		valType = val.Type()
	}
	return types.Variable{
		Name:        block.Label(),
		Default:     val,
		Type:        valType,
		Description: optionalString(block, "description"),
		Nullable:    optionalBoolean(block, "nullable"),
		Sensitive:   optionalBoolean(block, "sensitive"),
	}
}
