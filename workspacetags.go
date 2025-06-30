package preview

import (
	"fmt"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/json"

	"github.com/coder/preview/types"
)

func workspaceTags(modules terraform.Modules, files map[string]*hcl.File) (types.TagBlocks, hcl.Diagnostics) {
	diags := make(hcl.Diagnostics, 0)
	tagBlocks := make(types.TagBlocks, 0)

	for _, mod := range modules {
		blocks := mod.GetDatasByType("coder_workspace_tags")
		for _, block := range blocks {
			evCtx := block.Context().Inner()

			tagsAttr := block.GetAttribute("tags")
			if tagsAttr.IsNil() {
				r := block.HCLBlock().Body.MissingItemRange()
				diags = diags.Append(&hcl.Diagnostic{
					Severity:    hcl.DiagError,
					Summary:     "Missing required argument",
					Detail:      `"tags" attribute is required by coder_workspace_tags blocks`,
					Subject:     &r,
					EvalContext: evCtx,
				})
				continue
			}

			tagsValue := tagsAttr.Value()
			if !tagsValue.Type().IsObjectType() {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Incorrect type for \"tags\" attribute",
					// TODO: better error message for types
					Detail:      fmt.Sprintf(`"tags" attribute must be an 'Object', but got %q`, tagsValue.Type().FriendlyName()),
					Subject:     &tagsAttr.HCLAttribute().NameRange,
					Context:     &tagsAttr.HCLAttribute().Range,
					Expression:  tagsAttr.HCLAttribute().Expr,
					EvalContext: block.Context().Inner(),
				})
				continue
			}

			var tags []types.Tag
			tagsValue.ForEachElement(func(key cty.Value, val cty.Value) (stop bool) {
				r := tagsAttr.HCLAttribute().Expr.Range()
				tag, tagDiag := newTag(&r, files, key, val)
				if tagDiag != nil {
					diags = diags.Append(tagDiag)
					return false
				}

				tags = append(tags, tag)

				return false
			})

			tagBlocks = append(tagBlocks, types.TagBlock{
				Tags:  tags,
				Block: block,
			})
		}
	}

	return tagBlocks, diags
}

// newTag creates a workspace tag from its hcl expression.
func newTag(srcRange *hcl.Range, _ map[string]*hcl.File, key, val cty.Value) (types.Tag, *hcl.Diagnostic) {
	if key.IsKnown() && key.Type() != cty.String {
		return types.Tag{}, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid key type for tags",
			Detail:   fmt.Sprintf("Key must be a string, but got %s", key.Type().FriendlyName()),
			Context:  srcRange,
		}
	}

	tag := types.Tag{
		Key: types.HCLString{
			Value: key,
		},
		Value: types.HCLString{
			Value: val,
		},
	}

	switch val.Type() {
	case cty.String, cty.Bool, cty.Number:
		// These types are supported and can be safely converted to a string.
	default:
		fr := "<nil>"
		if !val.Type().Equals(cty.NilType) {
			fr = val.Type().FriendlyName()
		}

		// Unsupported types will be converted to a JSON string representation.
		jsonData, err := json.Marshal(val, val.Type())
		if err != nil {
			tag.Value.ValueDiags = tag.Value.ValueDiags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Invalid value type for tag %q", tag.KeyString()),
				Detail:   fmt.Sprintf("Value must be a string, but got %s. Attempt to marshal to json: %s", fr, err.Error()),
				Context:  srcRange,
			})
		} else {
			// Value successfully marshaled to JSON, we can store it as a string.
			tag.Value.Value = cty.StringVal(string(jsonData))
		}
	}

	return tag, nil
}
