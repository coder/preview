package preview

import (
	"io/fs"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	tfcontext "github.com/aquasecurity/trivy/pkg/iac/terraform/context"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"golang.org/x/xerrors"
)

func workspaceOwnerHook(dfs fs.FS, input Input) (func(ctx *tfcontext.Context, blocks terraform.Blocks, inputVars map[string]cty.Value), error) {
	if input.Owner.Groups == nil {
		input.Owner.Groups = []string{}
	}
	ownerGroups, err := gocty.ToCtyValue(input.Owner.Groups, cty.List(cty.String))
	if err != nil {
		return nil, xerrors.Errorf("converting owner groups: %w", err)
	}

	ownerValue := cty.ObjectVal(map[string]cty.Value{
		"groups": ownerGroups,
	})

	return func(ctx *tfcontext.Context, blocks terraform.Blocks, inputVars map[string]cty.Value) {
		for _, block := range blocks.OfType("data") {
			// TODO: Does it have to be me?
			if block.TypeLabel() == "coder_workspace_owner" && block.NameLabel() == "me" {
				block.Context().Parent().Set(ownerValue,
					"data", block.TypeLabel(), block.NameLabel())
			}
		}
	}, nil
}
