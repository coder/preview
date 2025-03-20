package types

import (
	"encoding/json"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type NullHCLString struct {
	Value string `json:"value"`
	Valid bool   `json:"valid"`
}

// @typescript-ignore HCLString
type HCLString struct {
	Value cty.Value

	// ValueDiags are any diagnostics that occurred
	// while evaluating the value
	ValueDiags hcl.Diagnostics
	// ValueExp is the underlying source expression
	ValueExpr hcl.Expression
	// Source is the literal hcl text that was parsed.
	// This is a best effort, it may not be available.
	Source *string
}

func ToHCLString(block *terraform.Block, attr *terraform.Attribute) HCLString {
	val, diags := attr.HCLAttribute().Expr.Value(block.Context().Inner())

	return HCLString{
		Value:      val,
		ValueDiags: diags,
		ValueExpr:  attr.HCLAttribute().Expr,
		// ??
		//Source:     attr.HCLAttribute().Range,
	}
}

func (s HCLString) MarshalJSON() ([]byte, error) {
	return json.Marshal(NullHCLString{
		Value: s.AsString(),
		Valid: s.Valid() && s.Value.IsKnown(),
	})
}

func (s *HCLString) UnmarshalJSON(data []byte) error {
	var reduced NullHCLString
	if err := json.Unmarshal(data, &reduced); err != nil {
		return err
	}
	if reduced.Valid {
		*s = StringLiteral(reduced.Value)
	} else {
		s.Value = cty.NilVal
	}

	return nil
}

func StringLiteral(s string) HCLString {
	v := cty.StringVal(s)
	return HCLString{
		Value:     v,
		ValueExpr: &hclsyntax.LiteralValueExpr{Val: v},
	}
}

// AsString is a safe function. It will always return a string.
// The caller should check if this value is Valid and known before
// calling this function.
func (s HCLString) AsString() string {
	if s.Valid() && s.Value.IsKnown() {
		switch {
		case s.Value.Type().Equals(cty.String):
			return s.Value.AsString()
		case s.Value.Type().Equals(cty.Number):
			// TODO: Float vs Int?
			return s.Value.AsBigFloat().String()
		case s.Value.Type().Equals(cty.Bool):
			if s.Value.True() {
				return "true"
			}
			return "false"
		default:
			// ?? What to do?
		}
	}

	if s.Source != nil {
		return *s.Source
	}

	return "??"
}

func (s HCLString) IsKnown() bool {
	return s.Valid() && s.Value.IsWhollyKnown()
}

func (s HCLString) Valid() bool {
	if s.ValueDiags.HasErrors() {
		return false
	}

	// TODO: Terraform seems to automatically cast these into strings?
	if !(s.Value.Type().Equals(cty.String) ||
		s.Value.Type().Equals(cty.Number) ||
		s.Value.Type().Equals(cty.Bool)) {
		return false
	}

	if s.Value.IsNull() {
		return false
	}

	return true
}
