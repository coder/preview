package preview

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint/terraform"
)

type DataDef struct {
	Type   string
	Schema *hcl.BodySchema
}

type ParsedHCLBlock struct {
	Block   *hcl.Block
	Content *hcl.BodyContent
}

func DataBlocks(schm DataDef, eval *terraform.Evaluator, mod *terraform.Module) ([]ParsedHCLBlock, hcl.Diagnostics) {
	blocks := make([]ParsedHCLBlock, 0)
	var diags hcl.Diagnostics
	// TODO: override files matter
	for _, f := range mod.Files {
		expanded, d := eval.ExpandBlock(f.Body, &hclext.BodySchema{})
		diags = diags.Extend(d)

		cc, _, d := expanded.PartialContent(&hcl.BodySchema{
			Attributes: nil,
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type:       "data",
					LabelNames: []string{"type", "name"},
				},
			},
		})
		diags = diags.Extend(d)

		for _, b := range cc.Blocks {
			if len(b.Labels) != 2 || b.Labels[0] != schm.Type {
				continue
			}
			b := b

			tagc, d := b.Body.Content(schm.Schema)
			diags = diags.Extend(d)

			blocks = append(blocks, ParsedHCLBlock{
				Block:   b,
				Content: tagc,
			})
		}
	}

	return blocks, diags
}
