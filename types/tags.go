package types

import (
	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/coder/preview/hclext"
)

type TagBlocks []TagBlock

func (b TagBlocks) ValidTags() map[string]string {
	tags := make(map[string]string)
	for _, block := range b {
		for key, value := range block.ValidTags() {
			tags[key] = value
		}
	}
	return tags
}

func (b TagBlocks) InvalidTags() Tags {
	tags := make(Tags, 0)
	for _, block := range b {
		tags = append(tags, block.InvalidTags()...)
	}
	return tags
}

type TagBlock struct {
	Tags  Tags
	Block *terraform.Block
}

func (b TagBlock) InvalidTags() Tags {
	invalid := make(Tags, 0)
	for _, tag := range b.Tags {
		if tag.Valid() {
			continue
		}

		invalid = append(invalid, tag)
	}
	return invalid
}

func (b TagBlock) ValidTags() map[string]string {
	tags := make(map[string]string)
	for _, tag := range b.Tags {
		if !tag.Valid() {
			continue
		}

		tags[tag.Key.AsString()] = tag.Value.AsString()
	}
	return tags
}

type Tags []Tag

func (t Tags) SafeNames() []string {
	names := make([]string, 0)
	for _, tag := range t {
		names = append(names, tag.SafeKeyString())
	}
	return names
}

type Tag struct {
	Key cty.Value
	// SafeKeyID can be used if the Key val is unknown
	SafeKeyID string
	KeyExpr   hcl.Expression

	Value     cty.Value
	ValueExpr hcl.Expression
}

func (t Tag) Valid() bool {
	if !t.Key.IsWhollyKnown() || !t.Value.IsWhollyKnown() {
		return false
	}
	if !t.Key.Type().Equals(cty.String) || !t.Value.Type().Equals(cty.String) {
		return false
	}
	return true
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
