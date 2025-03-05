package types

import (
	"github.com/aquasecurity/trivy/pkg/iac/terraform"

	"github.com/coder/preview/hclext"
)

// @typescript-ignore TagBlocks
type TagBlocks []TagBlock

func (b TagBlocks) Tags() map[string]string {
	tags := make(map[string]string)
	for _, block := range b {
		for key, value := range block.ValidTags() {
			tags[key] = value
		}
	}
	return tags
}

func (b TagBlocks) UnusableTags() Tags {
	tags := make(Tags, 0)
	for _, block := range b {
		tags = append(tags, block.UnusableTags()...)
	}
	return tags
}

// @typescript-ignore TagBlock
type TagBlock struct {
	Tags  Tags
	Block *terraform.Block
}

func (b TagBlock) UnusableTags() Tags {
	invalid := make(Tags, 0)
	for _, tag := range b.Tags {
		if tag.Valid() && tag.IsKnown() {
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

// @typescript-ignore Tags
type Tags []Tag

func (t Tags) SafeNames() []string {
	names := make([]string, 0)
	for _, tag := range t {
		names = append(names, tag.KeyString())
	}
	return names
}

// @typescript-ignore Tag
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
