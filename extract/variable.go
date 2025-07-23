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

// VariableFromBlock extracts a terraform variable, but not it's final resolved value.
// code taken mostly from https://github.com/aquasecurity/trivy/blob/main/pkg/iac/scanners/terraform/parser/evaluator.go#L479
func VariableFromBlock(block *terraform.Block) types.Variable {
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
				Diagnostics: hcl.Diagnostics{&hcl.Diagnostic{
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

		valOK := !val.IsNull() && val.IsWhollyKnown()
		typedVal, err := convert.Convert(val, valType)
		if err != nil && valOK {
			return types.Variable{
				Name: block.Label(),
				Diagnostics: hcl.Diagnostics{&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary: fmt.Sprintf("Failed to convert variable default value to type %q for %q",
						valType.FriendlyNameForConstraint(), block.Label()),
					Detail:  err.Error(),
					Subject: &defSubject,
				}},
			}
		}

		// If the new converted value is ok, use it.
		if err == nil {
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
		Ephemeral:   optionalBoolean(block, "ephemeral"),
	}
}
