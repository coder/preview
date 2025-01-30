package types

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/coder/preview/hclext"
)

type TagBlocks []TagBlock

type TagBlock struct {
	Tags []Tag

	Block   *hcl.Block
	Content *hcl.BodyContent
}

type Tag struct {
	Key cty.Value
	// SafeKeyID can be used if the Key val is unknown
	SafeKeyID string
	KeyExpr   hcl.Expression

	Value     cty.Value
	ValueExpr hcl.Expression
}

func (t Tag) IsKnown() bool {
	return t.Key.IsKnown() && t.Value.IsKnown()
}

func (t Tag) AsStrings() (string, string) {
	return t.Key.AsString(), t.Value.AsString()
}

func (t Tag) References() []string {
	return append(hclext.ReferenceNames(t.KeyExpr), hclext.ReferenceNames(t.ValueExpr)...)
}

func (t Tag) SafeKeyString() string {
	if t.Key.Type().Equals(cty.String) {
		return t.Key.AsString()
	}
	return t.SafeKeyID
}
