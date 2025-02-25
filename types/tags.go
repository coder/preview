package types

import (
	"github.com/aquasecurity/trivy/pkg/iac/terraform"

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
		if !tag.Valid() || !tag.IsKnown() {
			continue
		}

		k, v := tag.AsStrings()
		tags[k] = v
	}
	return tags
}

type Tags []Tag

func (t Tags) SafeNames() []string {
	names := make([]string, 0)
	for _, tag := range t {
		names = append(names, tag.KeyString())
	}
	return names
}

type Tag struct {
	Key   HCLString
	Value HCLString
}

func (t Tag) Valid() bool {
	return t.Key.Valid() && t.Value.Valid()
}

func (t Tag) IsKnown() bool {
	return t.Key.IsKnown() && t.Value.IsKnown()
}

func (t Tag) KeyString() string {
	return t.Key.AsString()
}

func (t Tag) AsStrings() (string, string) {
	return t.KeyString(), t.Value.AsString()
}

func (t Tag) References() []string {
	return append(hclext.ReferenceNames(t.Key.ValueExpr), hclext.ReferenceNames(t.Value.ValueExpr)...)
}
