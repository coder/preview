package preview

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/zclconf/go-cty/cty"

	"github.com/coder/preview/types"
)

func workspaceTags(eval *terraform.Evaluator, mod *terraform.Module) (types.TagBlocks, hcl.Diagnostics) {
	var tagBlocks []types.TagBlock
	var diags hcl.Diagnostics

	blocks, diags := DataBlocks(DataDef{
		Type: "coder_workspace_tags",
		Schema: &hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{
				{
					Name:     "tags",
					Required: true,
				},
			},
			Blocks: nil,
		},
	}, eval, mod)
	if diags.HasErrors() {
		return nil, diags
	}

	for _, block := range blocks {
		tagsAttr := block.Content.Attributes["tags"]
		if tagsAttr == nil {
			//r := block.Body.MissingItemRange()
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Missing required argument",
				Detail:   `"tags" attribute is required by coder_workspace_tags blocks`,
				Subject:  &block.Content.MissingItemRange,
			})
			continue
		}

		tagObj, ok := tagsAttr.Expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			diags = diags.Append(&hcl.Diagnostic{
				Severity:   hcl.DiagError,
				Summary:    "Incorrect type for \"tags\" attribute",
				Detail:     fmt.Sprintf(`"tags" attribute must be an 'ObjectConsExpr', but got %T`, tagsAttr.Expr),
				Subject:    &tagsAttr.NameRange,
				Context:    &tagsAttr.Range,
				Expression: tagsAttr.Expr,
			})
			continue
		}

		var tags []types.Tag
		for _, item := range tagObj.Items {
			key, kdiags := eval.EvaluateExpr(item.KeyExpr, cty.String)
			val, vdiags := eval.EvaluateExpr(item.ValueExpr, cty.String)

			diags = diags.Extend(kdiags)
			diags = diags.Extend(vdiags)

			if kdiags.HasErrors() {
				key = cty.UnknownVal(cty.String)
			}

			if vdiags.HasErrors() {
				val = cty.UnknownVal(cty.NilType)
			}

			safe, err := Source(item.KeyExpr.Range(), mod)
			if err != nil {
				safe = []byte("???") // we could do more here
			}

			tags = append(tags, types.Tag{
				Key:       key,
				Value:     val,
				SafeKeyID: string(safe),
				KeyExpr:   item.KeyExpr,
				ValueExpr: item.ValueExpr,
			})
		}
		tagBlocks = append(tagBlocks, types.TagBlock{
			Tags:    tags,
			Block:   block.Block,
			Content: block.Content,
		})
	}

	return tagBlocks, diags
}
