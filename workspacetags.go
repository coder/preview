package preview

import (
	"fmt"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"

	"github.com/coder/preview/types"
)

func WorkspaceTags(modules terraform.Modules, files map[string]*hcl.File) (types.TagBlocks, hcl.Diagnostics) {
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

			tagsObj, ok := tagsAttr.HCLAttribute().Expr.(*hclsyntax.ObjectConsExpr)
			if !ok {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Incorrect type for \"tags\" attribute",
					// TODO: better error message for types
					Detail:      fmt.Sprintf(`"tags" attribute must be an 'ObjectConsExpr', but got %T`, tagsAttr.HCLAttribute().Expr),
					Subject:     &tagsAttr.HCLAttribute().NameRange,
					Context:     &tagsAttr.HCLAttribute().Range,
					Expression:  tagsAttr.HCLAttribute().Expr,
					EvalContext: block.Context().Inner(),
				})
				continue
			}

			var tags []types.Tag
			for _, item := range tagsObj.Items {
				key, kdiags := item.KeyExpr.Value(evCtx)
				val, vdiags := item.ValueExpr.Value(evCtx)

				// TODO: what do do with the diags?
				if kdiags.HasErrors() {
					key = cty.UnknownVal(cty.String)
				}
				if vdiags.HasErrors() {
					val = cty.UnknownVal(cty.String)
				}

				if key.IsKnown() && key.Type() != cty.String {
					r := item.KeyExpr.Range()
					diags = diags.Append(&hcl.Diagnostic{
						Severity:    hcl.DiagError,
						Summary:     "Invalid key type for tags",
						Detail:      fmt.Sprintf("Key must be a string, but got %s", key.Type().FriendlyName()),
						Subject:     &r,
						Context:     &tagsObj.SrcRange,
						Expression:  item.KeyExpr,
						EvalContext: evCtx,
					})
					continue
				}

				safe, err := source(item.KeyExpr.Range(), files)
				if err != nil {
					safe = []byte("???") // we could do more here
				}

				tags = append(tags, types.Tag{
					Key:       key,
					SafeKeyID: string(safe),
					KeyExpr:   item.KeyExpr,
					Value:     val,
					ValueExpr: item.ValueExpr,
				})
			}
			tagBlocks = append(tagBlocks, types.TagBlock{
				Tags:  tags,
				Block: block,
			})
		}
	}

	return tagBlocks, diags
}
