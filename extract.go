package preview

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint/terraform"

	"github.com/coder/preview/types"
)

func extract(eval *terraform.Evaluator, mod *terraform.Module, input Input) (Output, hcl.Diagnostics) {
	//pcDiags := ParameterContexts(modules, input)
	tags, tagDiags := workspaceTags(eval, mod)
	params, rpDiags := richParameters(eval, mod)

	return Output{
		WorkspaceTags: tags,
		Parameters:    params,
		Files:         mod.Files,
	}, tagDiags.Extend(rpDiags)
}

func richParameters(eval *terraform.Evaluator, mod *terraform.Module) ([]types.Parameter, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	richParameters2(eval, mod)

	blocks, diags := DataBlocks(DataDef{
		Type: "coder_parameter",
		Schema: &hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type: "option",
				},
			},
			Attributes: []hcl.AttributeSchema{
				{
					Name:     "name",
					Required: true,
				},
				{
					Name:     "type",
					Required: false,
				},
				{
					Name:     "description",
					Required: false,
				},
				{
					Name:     "default",
					Required: false,
				},
			},
		},
	}, eval, mod)

	rps := make([]types.Parameter, 0)
	for _, block := range blocks {
		p := newAttributeParser(block.Content, eval)
		sch := &hclext.BodySchema{
			Attributes: []hclext.AttributeSchema{
				{
					Name:     "name",
					Required: true,
				},
				{
					Name:     "value",
					Required: true,
				},
			},
			Blocks: []hclext.BlockSchema{
				{
					Type:       "option",
					LabelNames: nil,
				},
			},
		}
		bb, bbdiags := eval.ExpandBlock(block.Block.Body, sch)
		var _ = bbdiags

		options, diag := bb.Content(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{},
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type:       "option",
					LabelNames: nil,
				},
			},
		})
		if diag.HasErrors() {
			diags = diags.Extend(diag)
			continue
		}

		var _ = options

		rp := types.Parameter{
			RichParameter: types.RichParameter{
				Name:         p.Attr("name").required().string(),
				Description:  p.Attr("description").required().string(),
				Type:         p.Attr("type").required().string(),
				Mutable:      false,
				DefaultValue: p.Attr("description").required().string(),
				Icon:         "",
				Options:      []*types.RichParameterOption{},
				Validation:   nil,
				Required:     false,
				DisplayName:  "",
				Order:        0,
				Ephemeral:    false,
			},
		}

		if p.diags.HasErrors() {
			diags = diags.Extend(p.diags)
			continue
		}

		rps = append(rps, rp)
	}

	return rps, diags
}

//func richParameterOptions(eval *terraform.Evaluator, bc *hcl.BodyContent) ([]types.RichParameterOption, hcl.Diagnostics) {
//	rpos := make([]types.RichParameterOption, 0)
//	for _, b := range bc.Blocks {
//		p := newAttributeParser(b, eval)
//		opt := types.RichParameterOption{}
//	}
//}
