package types

import (
	"encoding/json"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

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

func (s HCLString) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.AsString())
}

func (s *HCLString) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	*s = StringLiteral(str)
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
		return s.Value.AsString()
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

	if !s.Value.Type().Equals(cty.String) {
		return false
	}

	if s.Value.IsNull() {
		return false
	}

	return true
}
