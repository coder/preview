package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	ctyjson "github.com/zclconf/go-cty/cty/json"

	"github.com/coder/preview"
	"github.com/coder/preview/cli/clidisplay"
	"github.com/coder/preview/types"
	"github.com/coder/serpent"
)

type RootCmd struct {
	Files map[string]*hcl.File
}

func (r *RootCmd) Root() *serpent.Command {
	var (
		dir      string
		vars     []string
		groups   []string
		planJSON string
		preset   string
	)
	cmd := &serpent.Command{
		Use:   "codertf",
		Short: "codertf is a command line tool for previewing terraform template outputs.",
		Options: serpent.OptionSet{
			{
				Name:          "dir",
				Description:   "Directory with terraform files.",
				Flag:          "dir",
				FlagShorthand: "d",
				Default:       ".",
				Value:         serpent.StringOf(&dir),
			},
			{
				Name:          "plan",
				Description:   "Terraform plan file as json.",
				Flag:          "plan",
				FlagShorthand: "p",
				Default:       "",
				Value:         serpent.StringOf(&planJSON),
			},
			{
				Name:          "vars",
				Description:   "Variables.",
				Flag:          "vars",
				FlagShorthand: "v",
				Default:       "",
				Value:         serpent.StringArrayOf(&vars),
			},
			{
				Name:          "groups",
				Description:   "Groups.",
				Flag:          "groups",
				FlagShorthand: "g",
				Default:       "",
				Value:         serpent.StringArrayOf(&groups),
			},
			{
				Name:          "preset",
				Description:   "Name of the preset to define parameters. Run preview without this flag first to see a list of presets.",
				Flag:          "preset",
				FlagShorthand: "s",
				Default:       "",
				Value:         serpent.StringOf(&preset),
			},
		},
		Handler: func(i *serpent.Invocation) error {
			dfs := os.DirFS(dir)

			ctx := i.Context()

			output, _ := preview.Preview(ctx, preview.Input{}, dfs)
			presets := output.Presets
			chosenPresetIndex := slices.IndexFunc(presets, func(p types.Preset) bool {
				return p.Name == preset
			})

			rvars := make(map[string]string)
			for _, val := range vars {
				parts := strings.Split(val, "=")
				if len(parts) != 2 {
					continue
				}
				rvars[parts[0]] = parts[1]
			}
			if chosenPresetIndex != -1 {
				for paramName, paramValue := range presets[chosenPresetIndex].Parameters {
					rvars[paramName] = paramValue
				}
			}

			input := preview.Input{
				PlanJSONPath:    planJSON,
				ParameterValues: rvars,
				Owner: types.WorkspaceOwner{
					Groups: groups,
				},
			}

			output, diags := preview.Preview(ctx, input, dfs)
			if output == nil {
				return diags
			}
			r.Files = output.Files

			if len(diags) > 0 {
				_, _ = fmt.Fprintf(os.Stderr, "Parsing Diagnostics:\n")
				clidisplay.WriteDiagnostics(os.Stderr, output.Files, diags)
			}

			diags = clidisplay.WorkspaceTags(os.Stdout, output.WorkspaceTags)
			if len(diags) > 0 {
				_, _ = fmt.Fprintf(os.Stderr, "Workspace Tags Diagnostics:\n")
				clidisplay.WriteDiagnostics(os.Stderr, output.Files, diags)
			}

			if chosenPresetIndex == -1 {
				clidisplay.Presets(os.Stdout, presets, output.Files)
			}

			clidisplay.Parameters(os.Stdout, output.Parameters, output.Files)

			if !output.ModuleOutput.IsNull() && !(output.ModuleOutput.Type().IsObjectType() && output.ModuleOutput.LengthInt() == 0) {
				_, _ = fmt.Println("Module output")
				data, _ := ctyjson.Marshal(output.ModuleOutput, output.ModuleOutput.Type())
				var buf bytes.Buffer
				_ = json.Indent(&buf, data, "", "  ")
				_, _ = fmt.Println(buf.String())
			}

			return nil
		},
	}
	cmd.AddSubcommands(r.TerraformPlan())
	cmd.AddSubcommands(r.WebsocketServer())
	cmd.AddSubcommands(r.SetEnv())
	return cmd
}

//nolint:unused
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
