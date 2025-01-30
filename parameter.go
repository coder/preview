package preview

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/zclconf/go-cty/cty"

	"github.com/coder/preview/types"
)

var (
	parameterSchema = &hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{Type: "data", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type: "option",
						Body: &hclext.BodySchema{
							Mode: hclext.SchemaJustAttributesMode,
						},
					},
				},
				Attributes: []hclext.AttributeSchema{
					{Name: "name"},
				},
			},
			},
		},
	}
)

func richParameters2(eval *terraform.Evaluator, mod *terraform.Module) ([]types.Parameter, hcl.Diagnostics) {
	content := &hclext.BodyContent{}
	diags := hcl.Diagnostics{}

	//for _, f := range mod.Files {
	//body, expdiags := eval.ExpandBlock(f.Body, parameterSchema)
	//diags = diags.Extend(expdiags)

	v, dd := eval.EvaluateExpr(hclExpr("data.coder_parameter.project.default"), cty.List(cty.Object(map[string]cty.Type{
		"value": cty.String,
		"name":  cty.String,
	})))
	fmt.Println(v, dd)

	c, d := mod.PartialContent(parameterSchema, eval)
	diags = diags.Extend(d)
	for name, attr := range c.Attributes {
		content.Attributes[name] = attr
	}

	for _, b := range c.Blocks {
		b := b
		if b.Labels[0] != "coder_parameter" {
			continue
		}
		content.Blocks = append(content.Blocks, b)
	}
	//}

	if diags.HasErrors() {
		return nil, diags
	}
	return nil, nil
}

func hclExpr(expr string) hcl.Expression {
	file, diags := hclsyntax.ParseConfig([]byte(fmt.Sprintf(`expr = %s`, expr)), "test.tf", hcl.InitialPos)
	if diags.HasErrors() {
		panic(diags)
	}
	attributes, diags := file.Body.JustAttributes()
	if diags.HasErrors() {
		panic(diags)
	}
	return attributes["expr"].Expr
}
