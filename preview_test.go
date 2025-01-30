package preview_test

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/require"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"

	"github.com/coder/preview"
	"github.com/coder/preview/display"
	"github.com/coder/preview/types"
)

//go:embed testdata
var testdata embed.FS

type TestThing struct {
	Name   string `hcl:"name,attr"`
	Nested Foo    `hcl:"nested,block"`
}

type Foo struct {
	Bar string `hcl:"bar,attr"`
}

func TestPreview(t *testing.T) {
	//root := "testdata"
	//entries, err := testdata.ReadDir(root)
	//require.NoError(t, err)

	//schema, part := gohcl.ImpliedBodySchema(&TestThing{})
	//fmt.Println(schema, part)

	sub, err := fs.Sub(testdata, "testdata/conditional")
	require.NoError(t, err)

	ctx := context.Background()
	output, diags := preview.Preview(ctx, preview.Input{
		ParameterValues: []types.ParameterValue{},
	}, sub)

	require.False(t, diags.HasErrors())

	var out bytes.Buffer
	wtDiags := display.WorkspaceTags(&out, output.WorkspaceTags)
	t.Log("\n" + out.String())
	require.False(t, wtDiags.HasErrors())
}

func TestExampleContent(t *testing.T) {
	src := `
noodle "foo" "bar" {
	type = "rice"

	bread "baz" {
		type  = "focaccia"
		baked = true
	}
	bread "quz" {
		type = "rye"
	}
}
variable "regions" {
  type    = set(string)
  default = ["us", "eu", "au"]
}

data "coder_parameter" "region" {
  name        = "Region"
  description = "Which region would you like to deploy to?"
  type        = "string"
  default     = tolist(var.regions)[0]
  
  dynamic "option" {
    for_each = var.regions
    content {
      name  = option.value
      value = option.value
    }
  }
}

`
	file, diags := hclsyntax.ParseConfig([]byte(src), "test.tf", hcl.InitialPos)
	if diags.HasErrors() {
		panic(diags)
	}

	//body, diags := hclext.Content(file.Body, &hclext.BodySchema{
	body, diags := hclext.PartialContent(file.Body, &hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "variable",
				LabelNames: []string{"name"},
			},
			{
				Type:       "data",
				LabelNames: []string{"type", "name"},
			},
			{
				Type:       "noodle",
				LabelNames: []string{"name", "subname"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "type"}},
					Blocks: []hclext.BlockSchema{
						{
							Type:       "bread",
							LabelNames: []string{"name"},
							Body: &hclext.BodySchema{
								Attributes: []hclext.AttributeSchema{
									{Name: "type", Required: true},
									{Name: "baked"},
								},
							},
						},
					},
				},
			},
		},
	})
	if diags.HasErrors() {
		panic(diags)
	}

	for i, noodle := range body.Blocks {
		fmt.Printf("- noodle[%d]: labels=%s, attributes=%d\n", i, noodle.Labels, len(noodle.Body.Attributes))
		for i, bread := range noodle.Body.Blocks {
			fmt.Printf("  - bread[%d]: labels=%s, attributes=%d\n", i, bread.Labels, len(bread.Body.Attributes))
		}
	}
	// Output:
	// - noodle[0]: labels=[foo bar], attributes=0
	//   - bread[0]: labels=[baz], attributes=1
	//   - bread[1]: labels=[quz], attributes=1
}
