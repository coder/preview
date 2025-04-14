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

	ownerType, err := gocty.ImpliedType(input.Owner)
	if err != nil {
		return nil, xerrors.Errorf("getting owner cty type: %w", err)
	}

	ownerValues, err := gocty.ToCtyValue(input.Owner, ownerType)
	if err != nil {
		return nil, xerrors.Errorf("converting owner context to cty: %w", err)
	}

	return func(ctx *tfcontext.Context, blocks terraform.Blocks, inputVars map[string]cty.Value) {
		for _, block := range blocks.OfType("data") {
			// TODO: Does it have to be me?
			if block.TypeLabel() == "coder_workspace_owner" && block.NameLabel() == "me" {
				block.Context().Parent().Set(ownerValues,
					"data", block.TypeLabel(), block.NameLabel())
			}
		}
	}, nil
}
