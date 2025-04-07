package extract

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"

	"github.com/coder/preview/hclext"
	"github.com/coder/preview/types"
	"github.com/coder/terraform-provider-coder/v2/provider"
)

func ParameterFromBlock(block *terraform.Block) (*types.Parameter, hcl.Diagnostics) {
	diags := required(block, "name")
	if diags.HasErrors() {
		return nil, diags
	}

	pType, typDiag := optionalStringEnum[types.ParameterType](block, "type", types.ParameterTypeString, func(s types.ParameterType) error {
		return s.Valid()
	})
	if typDiag != nil {
		diags = diags.Append(typDiag)
	}

	formType, formTypeDiags := optionalStringEnum[provider.ParameterFormType](block, "form_type", provider.ParameterFormTypeDefault, func(s provider.ParameterFormType) error {
		if !slices.Contains(provider.ParameterFormTypes(), s) {
			return fmt.Errorf("invalid form type %q, expected one of [%s]", s, strings.Join(toStrings(provider.ParameterFormTypes()), ", "))
		}
		return nil
	})
	if formTypeDiags != nil {
		diags = diags.Append(formTypeDiags)
	}

	pName, nameDiag := requiredString(block, "name")
	if nameDiag != nil {
		diags = diags.Append(nameDiag)
	}

	if diags.HasErrors() {
		return nil, diags
	}

	pVal := richParameterValue(block)

	def := types.StringLiteral("")
	defAttr := block.GetAttribute("default")
	if !defAttr.IsNil() {
		def = types.ToHCLString(block, defAttr)
	}

	ftmeta := optionalString(block, "form_type_metadata")
	formTypeMeta := make(map[string]any)
	if ftmeta != "" {
		_ = json.Unmarshal([]byte(ftmeta), &formTypeMeta)
	}

	p := types.Parameter{
		Value: pVal,
		ParameterData: types.ParameterData{
			Name:             pName,
			Description:      optionalString(block, "description"),
			Type:             pType,
			FormType:         formType,
			FormTypeMetadata: formTypeMeta,
			Mutable:          optionalBoolean(block, "mutable"),
			// Default value is always written as a string, then converted
			// to the correct type.
			DefaultValue: def,
			Icon:         optionalString(block, "icon"),
			Options:      make([]*types.ParameterOption, 0),
			Validations:  make([]*types.ParameterValidation, 0),
			Required:     optionalBoolean(block, "required"),
			DisplayName:  optionalString(block, "display_name"),
			Order:        optionalInteger(block, "order"),
			Ephemeral:    optionalBoolean(block, "ephemeral"),

			Source: block,
		},
	}

	optBlocks := block.GetBlocks("option")

	optionType, newFormType, err := provider.ValidateFormType(provider.OptionType(p.Type), len(optBlocks), p.FormType)
	var _ = optionType // TODO: Should we enforce this anywhere?
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Invalid parameter `type=%q` and `form_type=%q`", p.Type, p.FormType),
			Detail:   err.Error(),
			Context:  &block.HCLBlock().DefRange,
		})

		// Parameter cannot be used
		p.FormType = provider.ParameterFormTypeError
	} else {
		p.FormType = newFormType
	}

	for _, b := range optBlocks {
		opt, optDiags := ParameterOptionFromBlock(b)
		diags = diags.Extend(optDiags)

		if optDiags.HasErrors() {
			continue
		}

		p.Options = append(p.Options, &opt)
	}

	validBlocks := block.GetBlocks("validation")
	if len(validBlocks) > 1 {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Multiple 'validation' blocks found",
			Detail:   "Only one validation block is allowed",
			Subject:  &validBlocks[0].HCLBlock().TypeRange,
			Context:  &validBlocks[0].HCLBlock().DefRange,
		})
	}

	for _, b := range block.GetBlocks("validation") {
		// TODO: Only parse if only 1 valid block exists
		valid, validDiags := ParameterValidationFromBlock(b)
		diags = diags.Extend(validDiags)

		if validDiags.HasErrors() {
			continue
		}

		p.Validations = append(p.Validations, &valid)
	}

	ctyType, err := p.CtyType()
	if err != nil {
		paramTypeDiag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Invalid parameter type %q", p.Type),
			Detail:   err.Error(),
			Context:  &block.HCLBlock().DefRange,
		}

		if attr := block.GetAttribute("type"); attr != nil && !attr.IsNil() {
			paramTypeDiag.Subject = &attr.HCLAttribute().Range
			paramTypeDiag.Expression = attr.HCLAttribute().Expr
			paramTypeDiag.EvalContext = block.Context().Inner()
		}
		diags = diags.Append(paramTypeDiag)
		p.FormType = provider.ParameterFormTypeError
	}

	if ctyType != cty.NilType && pVal.Value.Type().Equals(cty.String) {
		// TODO: Wish we could support more types, but only string types are
		// allowed.
		valStr := pVal.Value.AsString()
		// Apply validations to the parameter value
		for _, v := range p.Validations {
			if err := v.Valid(string(pType), valStr); err != nil {
				diags = diags.Append(&hcl.Diagnostic{
					Severity:   hcl.DiagError,
					Summary:    fmt.Sprintf("Paramater validation failed for value %q", valStr),
					Detail:     err.Error(),
					Expression: pVal.ValueExpr,
				})
			}
		}
	}

	usageDiags := ParameterUsageDiagnostics(p)
	if usageDiags.HasErrors() {
		p.FormType = provider.ParameterFormTypeError
	}
	diags = diags.Extend(usageDiags)

	// Diagnostics are scoped to the parameter
	p.Diagnostics = types.Diagnostics(diags)

	return &p, nil
}

func ParameterUsageDiagnostics(p types.Parameter) hcl.Diagnostics {
	valErr := "The value of a parameter is required to be sourced (default or input) for the parameter to function."
	var diags hcl.Diagnostics
	if !p.Value.Valid() {
		diags = diags.Append(&hcl.Diagnostic{
			Severity:   hcl.DiagError,
			Summary:    "Parameter value is not valid",
			Detail:     valErr,
			Expression: p.Value.ValueExpr,
		})
	} else if !p.Value.IsKnown() {
		diags = diags.Append(&hcl.Diagnostic{
			Severity:   hcl.DiagError,
			Summary:    "Parameter value is unknown, it likely includes a reference without a value",
			Detail:     valErr,
			Expression: p.Value.ValueExpr,
		})
	}

	var badOpts int
	for _, opt := range p.Options {
		if !opt.Value.IsKnown() || !opt.Value.Valid() {
			badOpts++
		}
	}

	if badOpts > 0 {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Parameter contains %d invalid options", badOpts),
			Detail:   "The set of options cannot be resolved, and use of the parameter is limited.",
		})
	}

	return diags
}

func ParameterValidationFromBlock(block *terraform.Block) (types.ParameterValidation, hcl.Diagnostics) {
	diags := required(block, "error")
	if diags.HasErrors() {
		return types.ParameterValidation{}, diags
	}

	pErr, errDiag := requiredString(block, "error")
	if errDiag != nil {
		diags = diags.Append(errDiag)
	}

	if diags.HasErrors() {
		return types.ParameterValidation{}, diags
	}

	p := types.ParameterValidation{
		Regex:     nullableString(block, "regex"),
		Error:     pErr,
		Min:       nullableInteger(block, "min"),
		Max:       nullableInteger(block, "max"),
		Monotonic: nullableString(block, "monotonic"),
		Invalid:   nullableBoolean(block, "invalid"),
	}

	return p, diags
}

func ParameterOptionFromBlock(block *terraform.Block) (types.ParameterOption, hcl.Diagnostics) {
	diags := required(block, "name", "value")
	if diags.HasErrors() {
		return types.ParameterOption{}, diags
	}

	pName, nameDiag := requiredString(block, "name")
	if nameDiag != nil {
		diags = diags.Append(nameDiag)
	}

	valAttr := block.GetAttribute("value")

	if diags.HasErrors() {
		return types.ParameterOption{}, diags
	}

	p := types.ParameterOption{
		Name:        pName,
		Description: optionalString(block, "description"),
		Value:       types.ToHCLString(block, valAttr),
		Icon:        optionalString(block, "icon"),
	}

	return p, diags
}

func optionalStringEnum[T ~string](block *terraform.Block, key string, def T, valid func(s T) error) (T, *hcl.Diagnostic) {
	str := optionalString(block, key)
	if str == "" {
		return def, nil
	}

	if err := valid(T(str)); err != nil {
		tyAttr := block.GetAttribute(key)
		return "", &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Invalid %q attribute", key),
			Detail:   err.Error(),
			Subject:  &(tyAttr.HCLAttribute().Range),
			//Context:     &(block.HCLBlock().DefRange),
			Expression:  tyAttr.HCLAttribute().Expr,
			EvalContext: block.Context().Inner(),
		}
	}

	return T(str), nil
}

func requiredString(block *terraform.Block, key string) (string, *hcl.Diagnostic) {
	tyAttr := block.GetAttribute(key)
	tyVal := tyAttr.Value()
	if tyVal.Type() != cty.String {
		typeName := "<nil>"
		if !tyVal.Type().Equals(cty.NilType) {
			typeName = tyVal.Type().FriendlyName()
		}

		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Invalid %q attribute", key),
			Detail:   fmt.Sprintf("Expected a string, got %q", typeName),
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

func nullableBoolean(block *terraform.Block, key string) *bool {
	attr := block.GetAttribute(key)
	if attr == nil || attr.IsNil() {
		return nil
	}
	val := attr.Value()
	if val.Type() != cty.Bool {
		return nil
	}

	b := val.True()
	return &b
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

func nullableString(block *terraform.Block, key string) *string {
	attr := block.GetAttribute(key)
	if attr == nil || attr.IsNil() {
		return nil
	}
	val := attr.Value()
	if val.Type() != cty.String {
		return nil
	}

	str := val.AsString()
	return &str
}

func optionalString(block *terraform.Block, key string) string {
	attr := block.GetAttribute(key)
	if attr == nil || attr.IsNil() {
		return ""
	}
	val := attr.Value()
	if !val.Type().Equals(cty.String) {
		return ""
	}

	return val.AsString()
}

func required(block *terraform.Block, keys ...string) hcl.Diagnostics {
	var diags hcl.Diagnostics
	for _, key := range keys {
		attr := block.GetAttribute(key)
		value := cty.NilVal
		if attr != nil {
			value, _ = attr.HCLAttribute().Expr.Value(block.Context().Inner())
		}

		if attr == nil || attr.IsNil() || value == cty.NilVal {
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
	ref := block.Reference()
	travs := []hcl.Traverser{
		hcl.TraverseRoot{
			Name: "data",
		},
		hcl.TraverseAttr{
			Name: ref.TypeLabel(),
		},
		hcl.TraverseAttr{
			Name: ref.NameLabel(),
		},
	}

	raw := ref.RawKey()
	if !raw.IsNull() {
		travs = append(travs, hcl.TraverseIndex{
			Key:      raw,
			SrcRange: hcl.Range{},
		})
	}

	travs = append(travs, hcl.TraverseAttr{
		Name: "value",
	})

	valRef := hclsyntax.ScopeTraversalExpr{
		Traversal: travs,
	}

	val, diags := valRef.Value(block.Context().Inner())
	source := hclext.CreateDotReferenceFromTraversal(valRef.Traversal)
	return types.HCLString{
		Value:      val,
		ValueDiags: diags,
		ValueExpr:  &valRef,
		Source:     &source,
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

func toStrings[A ~string](l []A) []string {
	var r []string
	for _, v := range l {
		r = append(r, string(v))
	}
	return r
}
