package extract

import (
	"fmt"
	"strings"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/coder/preview/hclext"
	"github.com/coder/preview/types"
)

func ParameterFromBlock(block *terraform.Block) (*types.Parameter, hcl.Diagnostics) {
	diags := required(block, "name", "type")
	if diags.HasErrors() {
		return nil, diags
	}

	pType, typDiag := requiredString("type", block)
	if typDiag != nil {
		diags = diags.Append(typDiag)
	}

	pName, nameDiag := requiredString("name", block)
	if nameDiag != nil {
		diags = diags.Append(nameDiag)
	}

	if diags.HasErrors() {
		return nil, diags
	}

	pVal := richParameterValue(block)
	p := types.Parameter{
		Value: pVal,
		RichParameter: types.RichParameter{
			Name:        pName,
			Description: optionalString(block, "description"),
			Type:        pType,
			Mutable:     optionalBoolean(block, "mutable"),
			// Default value is always written as a string, then converted
			// to the correct type.
			DefaultValue: optionalString(block, "default"),
			Icon:         optionalString(block, "icon"),
			Options:      make([]*types.ParameterOption, 0),
			Validations:  make([]*types.ParameterValidation, 0),
			Required:     optionalBoolean(block, "required"),
			DisplayName:  optionalString(block, "display_name"),
			Order:        optionalInteger(block, "order"),
			Ephemeral:    optionalBoolean(block, "ephemeral"),
		},
	}

	for _, b := range block.GetBlocks("option") {
		opt, optDiags := ParameterOptionFromBlock(b)
		diags = diags.Extend(optDiags)

		if optDiags.HasErrors() {
			continue
		}

		p.Options = append(p.Options, &opt)
	}

	for _, b := range block.GetBlocks("validation") {
		valid, validDiags := ParameterValidationFromBlock(b)
		diags = diags.Extend(validDiags)

		if validDiags.HasErrors() {
			continue
		}

		p.Validations = append(p.Validations, &valid)
	}

	// Diagnostics are scoped to the parameter
	p.Diagnostics = diags

	return &p, nil
}

func ParameterValidationFromBlock(block *terraform.Block) (types.ParameterValidation, hcl.Diagnostics) {
	diags := required(block, "error")
	if diags.HasErrors() {
		return types.ParameterValidation{}, diags
	}

	pErr, errDiag := requiredString("error", block)
	if errDiag != nil {
		diags = diags.Append(errDiag)
	}

	if diags.HasErrors() {
		return types.ParameterValidation{}, diags
	}

	p := types.ParameterValidation{
		Regex:     optionalString(block, "regex"),
		Error:     pErr,
		Min:       nullableInteger(block, "min"),
		Max:       nullableInteger(block, "max"),
		Monotonic: optionalString(block, "monotonic"),
	}

	return p, diags
}

func ParameterOptionFromBlock(block *terraform.Block) (types.ParameterOption, hcl.Diagnostics) {
	diags := required(block, "name", "value")
	if diags.HasErrors() {
		return types.ParameterOption{}, diags
	}

	pName, nameDiag := requiredString("name", block)
	if nameDiag != nil {
		diags = diags.Append(nameDiag)
	}

	pVal, valDiag := requiredString("value", block)
	if valDiag != nil {
		diags = diags.Append(valDiag)
	}

	if diags.HasErrors() {
		return types.ParameterOption{}, diags
	}

	p := types.ParameterOption{
		Name:        pName,
		Description: optionalString(block, "description"),
		Value:       pVal,
		Icon:        optionalString(block, "icon"),
	}

	return p, diags
}

func requiredString(key string, block *terraform.Block) (string, *hcl.Diagnostic) {
	tyAttr := block.GetAttribute(key)
	tyVal := tyAttr.Value()
	if tyVal.Type() != cty.String {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Invalid %q attribute", key),
			Detail:   fmt.Sprintf("Expected a string, got %q", tyVal.Type().FriendlyName()),
			Subject:  &(tyAttr.HCLAttribute().Range),
			//Context:     &(block.HCLBlock().DefRange),
			Expression:  tyAttr.HCLAttribute().Expr,
			EvalContext: block.Context().Inner(),
		}

		if !tyVal.IsWhollyKnown() {
			refs := hclext.ReferenceNames(tyAttr.HCLAttribute().Expr)
			if len(refs) > 0 {
				diag.Detail = fmt.Sprintf("Value is not known, check the references [%s] are resolvable",
					strings.Join(refs, ", "))
			}
		}

		return "", diag
	}

	return tyVal.AsString(), nil
}

func optionalBoolean(block *terraform.Block, key string) bool {
	attr := block.GetAttribute(key)
	if attr == nil || attr.IsNil() {
		return false
	}
	val := attr.Value()
	if val.Type() != cty.Bool {
		return false
	}

	return val.True()
}

func nullableInteger(block *terraform.Block, key string) *int64 {
	attr := block.GetAttribute(key)
	if attr == nil || attr.IsNil() {
		return nil
	}
	val := attr.Value()
	if val.Type() != cty.Number {
		return nil
	}

	i, acc := val.AsBigFloat().Int64()
	var _ = acc // acc should be 0

	return &i
}

func optionalInteger(block *terraform.Block, key string) int64 {
	attr := block.GetAttribute(key)
	if attr == nil || attr.IsNil() {
		return 0
	}
	val := attr.Value()
	if val.Type() != cty.Number {
		return 0
	}

	i, acc := val.AsBigFloat().Int64()
	var _ = acc // acc should be 0

	return i
}

func optionalString(block *terraform.Block, key string) string {
	attr := block.GetAttribute(key)
	if attr == nil || attr.IsNil() {
		return ""
	}
	val := attr.Value()
	if val.Type() != cty.String {
		return ""
	}

	return val.AsString()
}

func required(block *terraform.Block, keys ...string) hcl.Diagnostics {
	var diags hcl.Diagnostics
	for _, key := range keys {
		attr := block.GetAttribute(key)
		if attr == nil || attr.IsNil() || attr.Value() == cty.NilVal {
			r := block.HCLBlock().Body.MissingItemRange()
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Missing required attribute %q", key),
				Detail:   fmt.Sprintf("The %s attribute is required", key),
				Subject:  &r,
				Extra:    nil,
			})
		}
	}
	return diags
}

func richParameterValue(block *terraform.Block) types.HCLString {
	// Find the value of the parameter from the context.
	paramPath := append([]string{"data"}, block.Labels()...)
	path := strings.Join(paramPath, ".")

	valueRef := hclext.ScopeTraversalExpr(append(paramPath, "value")...)
	val, diags := valueRef.Value(block.Context().Inner())
	return types.HCLString{
		Value:      val,
		ValueDiags: diags,
		ValueExpr:  &valueRef,
		Source:     &path,
	}
}

func ParameterCtyType(typ string) (cty.Type, error) {
	switch typ {
	case "string":
		return cty.String, nil
	case "number":
		return cty.Number, nil
	case "bool":
		return cty.Bool, nil
	case "list(string)":
		return cty.List(cty.String), nil
	default:
		return cty.Type{}, fmt.Errorf("unsupported type: %q", typ)
	}
}
