package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"

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
		},
		Handler: func(i *serpent.Invocation) error {
			dfs := os.DirFS(dir)

			rvars := make(map[string]string)
			for _, val := range vars {
				parts := strings.Split(val, "=")
				if len(parts) != 2 {
					continue
				}
				rvars[parts[0]] = parts[1]
			}

			input := preview.Input{
				PlanJSONPath:    planJSON,
				ParameterValues: rvars,
				Owner: types.WorkspaceOwner{
					Groups: groups,
				},
			}

			ctx := i.Context()
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

			clidisplay.Parameters(os.Stdout, output.Parameters, output.Files)

			return nil
		},
	}
	cmd.AddSubcommands(r.TerraformPlan())
	cmd.AddSubcommands(r.WebsocketServer())
	return cmd
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
